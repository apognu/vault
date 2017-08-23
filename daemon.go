package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"

	"github.com/apognu/vault/crypt"
	"github.com/apognu/vault/util"

	"path/filepath"

	"os"

	"fmt"

	"github.com/gorilla/mux"
)

type apiResponse struct {
	Message string       `json:"message,omitempty"`
	Secrets []string     `json:"secrets,omitempty"`
	Secret  *util.Secret `json:"secret,omitempty"`
}

var apiKey string

func apiKeyMiddleware(w http.ResponseWriter, r *http.Request, nxt http.HandlerFunc) {
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(apiKey), nil
	})

	if err != nil || !token.Valid {
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
			writeError(w, http.StatusUnauthorized, "could not verify token signature")
			return
		}
	}

	nxt(w, r)
}

func StartServer(listen *net.TCPAddr, key string) {
	apiKey = key

	r := mux.NewRouter()
	n := negroni.New()
	n.UseFunc(apiKeyMiddleware)

	r.HandleFunc("/", listHandler)
	r.HandleFunc("/{name:.*}", secretHandler)

	n.UseHandler(r)
	http.ListenAndServe(listen.String(), n)
}

func writeError(w http.ResponseWriter, code int, msg string) {
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
