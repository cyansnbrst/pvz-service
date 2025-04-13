package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/server"
)

type ScenariosTestSuite struct {
	BaseTestSuite
}

func TestScenariosSuite(t *testing.T) {
	suite.Run(t, new(ScenariosTestSuite))
}

func (s *ScenariosTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
}

func (s *ScenariosTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *ScenariosTestSuite) TestFullPvzWorkflow() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	moderatorToken := s.Login(ts, "moderator")

	pvzReq := pvzapi.PostPvzJSONRequestBody{
		City: "Москва",
	}
	pvzBody, err := json.Marshal(pvzReq)
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/pvz", bytes.NewReader(pvzBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+moderatorToken)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var pvzResp pvzapi.PVZ
	err = json.NewDecoder(resp.Body).Decode(&pvzResp)
	s.Require().NoError(err)
	pvzID := pvzResp.Id

	employeeToken := s.Login(ts, "employee")

	receptionReq := pvzapi.PostReceptionsJSONRequestBody{
		PvzId: *pvzID,
	}
	receptionBody, err := json.Marshal(receptionReq)
	s.Require().NoError(err)

	req, err = http.NewRequest(http.MethodPost, ts.URL+"/receptions", bytes.NewReader(receptionBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+employeeToken)

	resp, err = http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusCreated, resp.StatusCode)

	for i := 0; i < 50; i++ {
		productReq := pvzapi.PostProductsJSONRequestBody{
			PvzId: *pvzID,
			Type:  "обувь",
		}
		productBody, err := json.Marshal(productReq)
		s.Require().NoError(err)

		req, err = http.NewRequest(http.MethodPost, ts.URL+"/products", bytes.NewReader(productBody))
		s.Require().NoError(err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+employeeToken)

		resp, err = http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusCreated, resp.StatusCode)
	}

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/pvz/%s/close_last_reception", ts.URL, pvzID.String()), nil)
	s.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+employeeToken)

	resp, err = http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
}
