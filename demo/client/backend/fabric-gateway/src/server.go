/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

TODO (eventually):
  eventually refactor the mock backend as it has the potential of wider
  usefulness (the same applies also to the fabric gateway).
  Auction-specific aspects;
   - the bridge change-code has auction in names (trivial to remove)
   - the "/api/getRegisteredUsers" and, in particular,
     "/api/clock_auction/getDefaultAuction", are auction-specific
   - processing of response

  PS: probably also worth moving the calls to __init & __setup as well
  as the unpacking of the payload objects, which are specific to FPC
  to chaincode/fpc_chaincode.go (or handle these calls for non-fpc
  in chaincode/go_chaincode.go such that actual go chaincode doesn't
   have to know about it?)
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

var flagPort string
var flagDebug bool

const ccName = "FPCAuction"
const channelName = "Mychannel"

const defaultMspId = "Org1MSP"
const defaultOrg = "org1"

type Config struct {
	ConnectionProfile string `json:"connection_profile_filename"`
	ChannelName       string `json:"channel_name"`
	ChaincodeName     string `json:"chaincode_name"`
	Wallet            string `json:"wallet"`
	AdminEnrollmentID string `json:"adminEnrollmentID"`
	EnrollmentSecret  string `json:"enrollmentSecret"`
	BackendPort       string `json:"backend_port"`
	Users             User   `json:"users"`
}

type User struct {
	Name  string `json:"userName"`
	Roll  string `json:"userRoll"`
	MSPID string `json:"MSPID"`
}

var sdk *fabsdk.FabricSDK

func main() {
	flag.StringVar(&flagPort, "port", "3000", "Port to listen on")
	flag.BoolVar(&flagDebug, "debug", false, "debug output")
	flag.Parse()

	// read config.json
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var gwConfig *Config
	json.Unmarshal(bytes, &gwConfig)
	// from config.json, take connection_profile.json

	sdk, err := fabsdk.New(config.FromFile(gwConfig.ConnectionProfile))
	if err != nil {
		fmt.Println(errors.WithMessage(err, "failed to create SDK"))
		os.Exit(-1)
	}
	defer sdk.Close()

	clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user), fabsdk.WithOrg(org))
	// client for interacting directly with the ledger
	ledger, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	// start web service
	startServer()
}

func startServer() {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "x-user")

	r := gin.Default()
	r.Use(cors.New(config))

	// JS-server
	// app.get('/',(function(req,res){
	// 	res.send('Welcome to Spectrum Auction by Elbonia Communication Commission - Enabled by FPC');
	// }));

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Spectrum Auction by Elbonia Communication Commission - Enabled by FPC")
	})

	// JS-server
	// //  different routes defined
	// app.use('/api/cc', ccRoute);
	// app.use('/api', apiRoute);
	// app.use('/api/clock_auction', clockauctionRoute);

	// ccRoute handles invoke and query
	r.GET("/api/cc", getState)
	// chaincode API
	r.POST("/api/cc/invoke", invoke)
	// note that using a MockStub there is no need to differentiate between query and invoke
	r.POST("/api/cc/query", query)

	// apiRoute looks like it only handles this one
	r.GET("/api/getRegisteredUsers", getAllUsers)

	// clockactionRoute handles getDefaultAction
	r.GET("/api/clock_auction/getDefaultAuction", getDefaultAuction)
	r.GET("/api/clock_auction/getAuctionDetails/:auctionId", getAuctionDetails)

	r.Run(":" + flagPort)
}

func getState(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "ledgerState")
}

func getAllUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, nil)
}

func getDefaultAuction(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, nil)
}

func getAuctionDetails(c *gin.Context) {
	resp := ResponseObject{
		Status: ResponseStatus{
			RC:      0,
			Message: "Ok",
		},
	}
	c.IndentedJSON(http.StatusOK, resp)
}

type Payload struct {
	Tx   string
	Args []string
}

// the JSON objects returned from FPC

type ResponseStatus struct {
	RC      int    `json:"rc"`
	Message string `json:"message"`
}

type ResponseObject struct {
	Status   ResponseStatus `json:"status"`
	Response interface{}    `json:"response"`
}

// Unmarshallers for above to ensure the fields exists ...
// (would be nice if there would be a tag 'json:required,...' or alike but alas
// despite 4 years of requests and discussion nothing such has materialized

func (status *ResponseStatus) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		RC      *int    `json:"rc"`
		Message *string `json:"message"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if required.RC == nil || required.Message == nil {
		err = fmt.Errorf("Required fields for ResponseStatus missing")
	} else {
		status.RC = *required.RC
		status.Message = *required.Message
	}
	return
}

func (response *ResponseObject) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		Status *ResponseStatus `json:"status"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if required.Status == nil {
		err = fmt.Errorf("Required fields for ResponseStatus missing")
	} else {
		response.Status = *required.Status
	}
	return
}

// Main invocation handling
func invoke(c *gin.Context) {
	c.Data(http.StatusOK, c.ContentType(), nil)
}

// Main invocation handling
func query(c *gin.Context) {
	c.Data(http.StatusOK, c.ContentType(), nil)
}

func createFPCResponse(res peer.Response) []byte {

	// NOTE: we (try to) return error even if the invocation get success back
	// but does not contain a response payload. According to the auction
	// specifications, all queries and transactions should return a response
	// object (even more specifically, an object which at the very least
	// contains a 'status' field)
	var fpcResponse []byte
	var errMsg *string = nil // nil means no error
	// we might get payload and response regardless of invocation success,
	// so try to decode in all cases
	if res.Payload != nil {

	} else {
		msg := fmt.Sprintf("No response payload received (status=%v/message=%v)",
			res.Status, res.Message)
		errMsg = &msg
	}

	if errMsg != nil {
		fpcResponseJson := ResponseObject{
			Status: ResponseStatus{
				RC:      499, // TODO (maybe): more specific explicit error codes?
				Message: *errMsg,
			},
			Response: fpcResponse,
		}
		fpcResponse, _ = json.Marshal(fpcResponseJson)
	}
	return fpcResponse
}
