package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/csdengh/cur_blank/db/mock"
	db "github.com/csdengh/cur_blank/db/sqlc"
	"github.com/csdengh/cur_blank/token"
	"github.com/csdengh/cur_blank/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransfer(t *testing.T) {
	account1 := randomAccount()
	account2 := randomAccount()
	account3 := randomAccount()

	amount := int64(2)

	account1.Currency = "USD"
	account2.Currency = "USD"
	account3.Currency = "CAD"

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(ms *mockdb.MockStore) {
				ms.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				ms.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(account2, nil)

				transReq := db.TransferTxParams{
					FromAccountId: account1.ID,
					ToAccountId:   account2.ID,
					Amount:        amount,
				}

				transRes := db.TransferTxResult{
					FromAccount: account1,
					ToAccount:   account2,
					Err:         nil,
				}
				ms.EXPECT().TransferTx(gomock.Any(), gomock.Eq(&transReq)).Times(1).Return(&transRes)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthenticate(t, request, tokenMaker, authorizationTypeBearer, account1.Owner, time.Minute)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
			},
		},
	}

	config, err := utils.GetConfig("../")
	require.NoError(t, err)

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			mctrl := gomock.NewController(t)
			ms := mockdb.NewMockStore(mctrl)

			tc.buildStubs(ms)

			s, err := NewServer(config, ms)
			require.NoError(t, err)

			recoder := httptest.NewRecorder()
			url := "/transfers"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			tc.setupAuth(t, req, s.tokenMaker)
			require.NoError(t, err)
			s.route.ServeHTTP(recoder, req)

			tc.checkResponse(recoder)
		})
	}

}
