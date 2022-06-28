package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/csdengh/cur_blank/db/mock"
	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name:      "ok",
			accountId: account.ID,
			buildStubs: func(ms *mockdb.MockStore) {
				ms.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
				bodyConfirm(t, recoder.Body, account)
			},
		},
		{
			name:      "notfound",
			accountId: account.ID,
			buildStubs: func(ms *mockdb.MockStore) {
				ms.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recoder.Code)
			},
		},
		{
			name:      "internalError",
			accountId: account.ID,
			buildStubs: func(ms *mockdb.MockStore) {
				ms.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recoder.Code)
			},
		},
		{
			name:      "idFormatErr",
			accountId: 0,
			buildStubs: func(ms *mockdb.MockStore) {
				ms.EXPECT().GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			mockctl := gomock.NewController(t)
			defer mockctl.Finish()

			ms := mockdb.NewMockStore(mockctl)
			tc.buildStubs(ms)
			s := NewServer(ms)

			recoder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			s.route.ServeHTTP(recoder, req)

			tc.checkResponse(t, recoder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       utils.RandomInt(2, 200),
		Owner:    utils.RandomOwner(),
		Balance:  utils.RandomInt(2, 200),
		Currency: utils.RandomCurrency(),
	}
}

func bodyConfirm(t *testing.T, actual *bytes.Buffer, expire db.Account) {
	data, err := ioutil.ReadAll(actual)
	require.NoError(t, err)

	var actualAcc db.Account
	err = json.Unmarshal(data, &actualAcc)
	require.NoError(t, err)
	require.Equal(t, expire, actualAcc)
}
