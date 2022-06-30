package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "github.com/csdengh/cur_blank/db/mock"
	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type userMatcher struct {
	pwd  string
	user db.CreateUserParams
}

func (e userMatcher) Matches(x interface{}) bool {
	req, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := utils.ConfirmPwd(req.HashedPassword, e.pwd)
	if err != nil {
		return false
	}

	req.HashedPassword = e.user.HashedPassword

	return reflect.DeepEqual(e.user, req)
}

func (e userMatcher) String() string {
	return fmt.Sprintf("is equal to %s (%v)", e.pwd, e.user)
}

func UserMatch(pwd string, user db.CreateUserParams) userMatcher {
	return userMatcher{pwd: pwd, user: user}
}

func TestCreateUser(t *testing.T) {
	pwd := "hahahaha"
	user, err := createRamdomUser(pwd)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"username":  user.Username,
				"password":  pwd,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(ms *mockdb.MockStore) {

				args := db.CreateUserParams{
					Username:       user.Username,
					HashedPassword: user.HashedPassword,
					FullName:       user.FullName,
					Email:          user.Email,
				}

				ms.EXPECT().CreateUser(gomock.Any(), UserMatch(pwd, args)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
				userConfirm(t, recoder.Body, user)
			},
		},
	}

	config, err := utils.GetConfig("../")
	require.NoError(t, err)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ms := mockdb.NewMockStore(ctrl)

			tc.buildStubs(ms)

			s,err := NewServer(config, ms)
			require.NoError(t, err)

			recoder := httptest.NewRecorder()
			url := "/users"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			s.route.ServeHTTP(recoder, req)

			tc.checkResponse(t, recoder)

		})
	}

}

func createRamdomUser(pwd string) (db.User, error) {
	hashpwd, err := utils.HashPassword(pwd)
	if err != nil {
		return db.User{}, err
	}
	return db.User{
		Username:       utils.RandomOwner(),
		HashedPassword: hashpwd,
		FullName:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
	}, nil
}

func userConfirm(t *testing.T, actual *bytes.Buffer, expire db.User) {
	data, err := ioutil.ReadAll(actual)
	require.NoError(t, err)

	var actualAcc db.User
	err = json.Unmarshal(data, &actualAcc)
	require.NoError(t, err)
	expire.HashedPassword = ""
	require.Equal(t, expire, actualAcc)
}
