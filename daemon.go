package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"

	"github.com/apognu/vault/crypt"
	"github.com/apognu/vault/util"

	"path/filepath"

	"os"

	"fmt"

	"github.com/gorilla/mux"
)

type ResponseWriter struct {
	ResponseCode int
	Writer       http.ResponseWriter
}

func (w ResponseWriter) Header() http.Header {
	return w.Writer.Header()
}

func (w *ResponseWriter) WriteHeader(h int) {
	w.ResponseCode = h
	w.Writer.WriteHeader(h)
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type apiResponse struct {
	Message    string           `json:"message,omitempty"`
	Secrets    []string         `json:"secrets,omitempty"`
	Secret     *util.Secret     `json:"secret,omitempty"`
	MasterKeys []util.MasterKey `json:"master_keys,omitempty"`
}

var apiKey string

func apiKeyMiddleware(w http.ResponseWriter, r *http.Request, nxt http.HandlerFunc) {
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logrus.Errorf("unknown signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(apiKey), nil
	})

	if err != nil || !token.Valid {
    logrus.Errorf("could not validate token: %s", err.Error())
		writeError(w, http.StatusUnauthorized, "could not verify token signature")
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["exp"] == nil || claims["nbf"] == nil {
			writeError(w, http.StatusUnauthorized, "cannot find 'exp' and 'nbf' claims")
			return
		}
		exp, eok := claims["exp"].(float64)
		nbf, nok := claims["nbf"].(float64)
		if !eok || !nok {
			writeError(w, http.StatusUnauthorized, "failed to parse 'exp' and 'nbf' claims")
			return
		}
		if (exp - nbf) > 30 {
			writeError(w, http.StatusUnauthorized, "'exp' and 'nbf' should not be separated by more than 30 seconds")
			return
		}

		if claims["aud"] != fmt.Sprintf("vault:%s", r.URL.Path) {
      writeError(w, http.StatusUnauthorized, fmt.Sprintf("wrong audience: vault:%s != vault:%s", claims["aud"], r.URL.Path))
			return
		}
	}

	nxt(w, r)
}

func logMiddleware(w http.ResponseWriter, r *http.Request, nxt http.HandlerFunc) {
	wr := &ResponseWriter{ResponseCode: 200, Writer: w}
	t := float64(time.Now().UnixNano())

	nxt(wr, r)

	d := (float64(time.Now().UnixNano()) - t) / 1000000000
	url, _ := url.Parse(r.RequestURI)

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	log := logrus.Fields{
		"status":   wr.ResponseCode,
		"duration": fmt.Sprintf("%.3f", d),
	}

	logrus.WithFields(log).Infof("%s %s", r.Method, url.Path)
}

func StartServer(listen *net.TCPAddr, key string) {
	apiKey = key

	r := mux.NewRouter()
	n := negroni.New()
	n.UseFunc(logMiddleware)
	n.UseFunc(apiKeyMiddleware)

	r.HandleFunc("/", listHandler)
	r.HandleFunc("/-/masterkeys", masterKeysHandler)
	r.HandleFunc("/{name:.*}", secretHandler)

	n.UseHandler(r)
	http.ListenAndServe(listen.String(), n)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	logrus.Error(msg)

	j, _ := json.Marshal(apiResponse{Message: msg})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(j)
}

func writeResponse(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")

	j, err := json.Marshal(o)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not marshal response")
		return
	}

	w.Write(j)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	secrets := make([]string, 0)
	err := filepath.Walk(util.GetVaultPath(), func(path string, info os.FileInfo, err error) error {
		if walk, err := util.ShouldFileBeWalked(path); !walk {
			return err
		}

		secretPathTokens := strings.Split(path, util.GetVaultPath())
		if len(secretPathTokens) == 0 {
			return nil
		}

		secretPath := strings.Trim(secretPathTokens[1], "/")
		secrets = append(secrets, secretPath)

		return nil
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not get secrets")
	}

	writeResponse(w, apiResponse{Secrets: secrets})
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	secret, err := crypt.GetSecretFile(mux.Vars(r)["name"])
	if err != nil {
		writeError(w, http.StatusNotFound, "could not open secret file")
		return
	}

	writeResponse(w, apiResponse{Secret: secret})
}

func masterKeysHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, apiResponse{MasterKeys: crypt.GetVaultMeta(false).MasterKeys})
}
