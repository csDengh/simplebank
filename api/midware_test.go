package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/csdengh/cur_blank/token"
	"github.com/csdengh/cur_blank/utils"
	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/require"
)

func addAuthenticate(t *testing.T, request *http.Request, tokenMaker token.Maker, authorizationType string, username string, timedur time.Duration) {
	s, pl, err := tokenMaker.CreateToken(username, timedur)
	require.NoError(t, err)
	require.NotEmpty(t, pl)
	require.NotEmpty(t, s)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, s)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestMidware(t *testing.T) {
	testcases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthenticate(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthenticate(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthenticate(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthenticate(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	config, err := utils.GetConfig("../")
	require.NoError(t, err)

	for i := range testcases {
		tc := testcases[i]

		t.Run(tc.name, func(t *testing.T) {
			s, err := NewServer(config, nil)
			require.NoError(t, err)
			require.NotEmpty(t, s)

			authPath := "/auth"
			s.route.GET(authPath, AuthenticateMideware(s.tokenMaker), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			})

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, s.tokenMaker)
			s.route.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
