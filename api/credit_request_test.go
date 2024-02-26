package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/40grivenprog/simple-bank/db/mock"
	db "github.com/40grivenprog/simple-bank/db/sqlc"
	"github.com/40grivenprog/simple-bank/token"
	"github.com/40grivenprog/simple-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateCreditRequestAPI(t *testing.T) {
	user, _ := randomUser(t)
	creditRequest := randomCreditRequst(user.Username)
	account := db.Account{
		ID:       util.RandomInt(1, 100),
		Owner:    user.Username,
		Balance:  util.RandomMoney(),
		Currency: creditRequest.Currency,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"currency": creditRequest.Currency,
				"amount":   creditRequest.Amount,
				"reason":   creditRequest.Reason.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getArg := db.GetAccountByUsernameAndCurrencyParams{
					Owner:    user.Username,
					Currency: creditRequest.Currency,
				}
				store.EXPECT().
					GetAccountByUsernameAndCurrency(gomock.Any(), gomock.Eq(getArg)).
					Times(1).
					Return(account, nil)
				createArg := db.CreateCreditRequestParams{
					Currency: creditRequest.Currency,
					Amount:   creditRequest.Amount,
					Reason:   creditRequest.Reason,
					Username: user.Username,
				}
				store.EXPECT().
					CreateCreditRequest(gomock.Any(), gomock.Eq(createArg)).
					Times(1).
					Return(creditRequest, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				fmt.Println("ERROR", recorder.Body)
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCreditRequest(t, recorder.Body, creditRequest)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"currency": creditRequest.Currency,
				"reason":   creditRequest.Reason.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Bad Request When User haven't fot suitable account",
			body: gin.H{
				"currency": creditRequest.Currency,
				"amount":   creditRequest.Amount,
				"reason":   creditRequest.Reason.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				getArg := db.GetAccountByUsernameAndCurrencyParams{
					Owner:    user.Username,
					Currency: creditRequest.Currency,
				}
				store.EXPECT().
					GetAccountByUsernameAndCurrency(gomock.Any(), gomock.Eq(getArg)).
					Times(1).
					Return(db.Account{}, fmt.Errorf("user haven't got account with currency: %s", creditRequest.Currency))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			body: gin.H{
				"currency": creditRequest.Currency,
				"amount":   creditRequest.Amount,
				"reason":   creditRequest.Reason.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			cntrl := gomock.NewController(t)
			store := mockdb.NewMockStore(cntrl)
			defer cntrl.Finish()

			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/credit_requests"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListCreditRequests(t *testing.T) {
	user, _ := randomUser(t)
	creditRequest := randomCreditRequst(user.Username)

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreditRequestsByUsername(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return([]db.CreditRequest{creditRequest}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCreditRequestList(t, recorder.Body, []db.CreditRequest{creditRequest})
			},
		},
		{
			name: "Unauthorized",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			cntrl := gomock.NewController(t)
			store := mockdb.NewMockStore(cntrl)
			defer cntrl.Finish()

			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/credit_requests"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListPendingCreditRequests(t *testing.T) {
	adminUser, _ := randomAdminUser(t)
	baseUser, _ := randomUser(t)
	creditRequest := randomPendingCreditRequest(baseUser.Username)

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, adminUser.Username, adminUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUsersPendingCreditRequests(gomock.Any()).
					Times(1).
					Return([]db.CreditRequest{creditRequest}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCreditRequestList(t, recorder.Body, []db.CreditRequest{creditRequest})
			},
		},
		{
			name: "Forbidden",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, baseUser.Username, baseUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			cntrl := gomock.NewController(t)
			store := mockdb.NewMockStore(cntrl)
			defer cntrl.Finish()

			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/admin/credit_requests"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestCancelPendingRequest(t *testing.T) {
	adminUser, _ := randomAdminUser(t)
	baseUser, _ := randomUser(t)
	creditRequest := randomPendingCreditRequest(baseUser.Username)

	testCases := []struct {
		name          string
		url           string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			url:  fmt.Sprintf("/admin/credit_requests/%d", creditRequest.ID),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, adminUser.Username, adminUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				cancelledCreditRequest := creditRequest
				cancelledCreditRequest.Status = db.CreditRequestsStatusCancelled
				store.EXPECT().
					GetPendingCreditRequestById(gomock.Any(), gomock.Eq(creditRequest.ID)).
					Times(1).
					Return(creditRequest, nil)
				store.EXPECT().
					CancelCreditRequestById(gomock.Any(), gomock.Eq(creditRequest.ID)).
					Times(1).
					Return(cancelledCreditRequest, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				expectedCreditRequest := creditRequest
				expectedCreditRequest.Status = db.CreditRequestsStatusCancelled
				requireBodyMatchCreditRequest(t, recorder.Body, expectedCreditRequest)
			},
		},
		{
			name: "Forbidden",
			url:  fmt.Sprintf("/admin/credit_requests/%d", creditRequest.ID),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, baseUser.Username, baseUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "Bad request Invalid ID",
			url:  fmt.Sprintf("/admin/credit_requests/%d", 0),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, adminUser.Username, adminUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Bad request Not Pending Credit Request",
			url:  fmt.Sprintf("/admin/credit_requests/%d", creditRequest.ID),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, adminUser.Username, adminUser.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPendingCreditRequestById(gomock.Any(), gomock.Eq(creditRequest.ID)).
					Times(1).
					Return(db.CreditRequest{}, errors.New("no rows"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			cntrl := gomock.NewController(t)
			store := mockdb.NewMockStore(cntrl)
			defer cntrl.Finish()

			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := tc.url

			request, err := http.NewRequest(http.MethodPatch, url, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomPendingCreditRequest(username string) db.CreditRequest {
	creditRequest := randomCreditRequst(username)
	creditRequest.Status = db.CreditRequestsStatusPending

	return creditRequest
}

func randomCreditRequst(username string) db.CreditRequest {
	return db.CreditRequest{
		ID:       util.RandomInt(1, 100),
		Username: username,
		Amount:   int32(util.RandomInt(1, 100)),
		Currency: "USD",
		Status:   db.CreditRequestsStatusPending,
		Reason:   sql.NullString{String: "Reason"},
	}
}

func requireBodyMatchCreditRequest(t *testing.T, body *bytes.Buffer, creditRequest db.CreditRequest) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotCreditRequest db.CreditRequest
	err = json.Unmarshal(data, &gotCreditRequest)
	require.NoError(t, err)
	require.Equal(t, creditRequest, gotCreditRequest)
}

func requireBodyMatchCreditRequestList(t *testing.T, body *bytes.Buffer, expectedCreditRequests []db.CreditRequest) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotCreditRequests []db.CreditRequest
	err = json.Unmarshal(data, &gotCreditRequests)
	require.NoError(t, err)
	require.Equal(t, expectedCreditRequests, gotCreditRequests)
}
