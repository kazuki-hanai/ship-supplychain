/*
SPDX-License-Identifier: Apache-2.0
*/

package shipping

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Shipping data
type Shipping struct {
	Type         string             `json:"objectType"`
	ItemSold     Item               `json:"item"`
	Seller       string             `json:"seller"`
	Orgs         []string           `json:"organizations"`
	PrivateBids  map[string]BidHash `json:"privateBids"`
	RevealedBids map[string]FullBid `json:"revealedBids"`
	Winner       string             `json:"winner"`
	Price        int                `json:"price"`
	Status       string             `json:"status"`
}

type Item struct {
	Name   string `json:"item"`
	Dest   string `json:"dest"`
	Weight int    `json:"org"`
	Days   int    `json:"days"`
}

// FullBid is the structure of a revealed bid
type FullBid struct {
	Type   string `json:"objectType"`
	Price  int    `json:"price"`
	Org    string `json:"org"`
	Bidder string `json:"bidder"`
}

// BidHash is the structure of a private bid
type BidHash struct {
	Org  string `json:"org"`
	Hash string `json:"hash"`
}

const bidKeyType = "bid"

// CreateShipping creates on shipping on the public channel. The identity that
// submits the transacion becomes the seller of the shipping
func (s *SmartContract) CreateShipping(ctx contractapi.TransactionContextInterface, shippingID string, itemName string, itemDest string, itemWeight int, itemDays int) error {

	// get ID of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// get org of submitting client
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// Create shipping
	bidders := make(map[string]BidHash)
	revealedBids := make(map[string]FullBid)

	shipping := Shipping{
		Type: "shipping",
		ItemSold: Item{
			Name:   itemName,
			Dest:   itemDest,
			Weight: itemWeight,
			Days:   itemDays,
		},
		Price:        100000000,
		Seller:       clientID,
		Orgs:         []string{clientOrgID},
		PrivateBids:  bidders,
		RevealedBids: revealedBids,
		Winner:       "",
		Status:       "open",
	}

	shippingJSON, err := json.Marshal(shipping)
	if err != nil {
		return err
	}

	// put shipping into state
	err = ctx.GetStub().PutState(shippingID, shippingJSON)
	if err != nil {
		return fmt.Errorf("failed to put shipping in public data: %v", err)
	}

	// set the seller of the shipping as an endorser
	err = setAssetStateBasedEndorsement(ctx, shippingID, clientOrgID)
	if err != nil {
		return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
	}

	return nil
}

// Bid is used to add a user's bid to the shipping. The bid is stored in the private
// data collection on the peer of the bidder's organization. The function returns
// the transaction ID so that users can identify and query their bid
func (s *SmartContract) Bid(ctx contractapi.TransactionContextInterface, shippingID string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	BidJSON, ok := transientMap["bid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// the bidder has to target their peer to store the bid
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Cannot store bid on this peer, not a member of this org: Error %v", err)
	}

	// the transaction ID is used as a unique index for the bid
	txID := ctx.GetStub().GetTxID()

	// create a composite key using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{shippingID, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, bidKey, BidJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input price into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

// SubmitBid is used by the bidder to add the hash of that bid stored in private data to the
// shipping. Note that this function alters the shipping in private state, and needs
// to meet the shipping endorsement policy. Transaction ID is used identify the bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, shippingID string, txID string) error {

	// get the MSP ID of the bidder's org
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// get the shipping from public state
	shipping, err := s.QueryShipping(ctx, shippingID)
	if err != nil {
		return fmt.Errorf("failed to get shipping from public state %v", err)
	}

	// the shipping needs to be open for users to add their bid
	Status := shipping.Status
	if Status != "open" {
		return fmt.Errorf("cannot join closed or ended shipping")
	}

	// get the inplicit collection name of bidder's org
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// use the transaction ID passed as a parameter to create composite bid key
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{shippingID, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// get the hash of the bid stored in private data collection
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid bash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	// store the hash along with the bidder's organization
	NewHash := BidHash{
		Org:  clientOrgID,
		Hash: fmt.Sprintf("%x", bidHash),
	}

	bidders := make(map[string]BidHash)
	bidders = shipping.PrivateBids
	bidders[bidKey] = NewHash
	shipping.PrivateBids = bidders

	// Add the bidding organization to the list of participating organizations if it is not already
	Orgs := shipping.Orgs
	if !(contains(Orgs, clientOrgID)) {
		newOrgs := append(Orgs, clientOrgID)
		shipping.Orgs = newOrgs

		err = addAssetStateBasedEndorsement(ctx, shippingID, clientOrgID)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}

	newShippingJSON, _ := json.Marshal(shipping)

	err = ctx.GetStub().PutState(shippingID, newShippingJSON)
	if err != nil {
		return fmt.Errorf("failed to update shipping: %v", err)
	}

	return nil
}

// RevealBid is used by a bidder to reveal their bid after the shipping is closed
func (s *SmartContract) RevealBid(ctx contractapi.TransactionContextInterface, shippingID string, txID string) error {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientBidJSON, ok := transientMap["bid"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}

	// get implicit collection name of organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// use transaction ID to create composit bid key
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{shippingID, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// get bid hash of bid if private bid on the public ledger
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid bash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	// get shipping from public state
	shipping, err := s.QueryShipping(ctx, shippingID)
	if err != nil {
		return fmt.Errorf("failed to get shipping from public state %v", err)
	}

	// Complete a series of three checks before we add the bid to the shipping

	// check 1: check that the shipping is closed. We cannot reveal a
	// bid to an open shipping
	Status := shipping.Status
	if Status != "closed" {
		return fmt.Errorf("cannot reveal bid for open or ended shipping")
	}

	// check 2: check that hash of revealed bid matches hash of private bid
	// on the public ledger. This checks that the bidder is telling the truth
	// about the value of their bid

	hash := sha256.New()
	hash.Write(transientBidJSON)
	calculatedBidJSONHash := hash.Sum(nil)

	// verify that the hash of the passed immutable properties matches the on-chain hash
	if !bytes.Equal(calculatedBidJSONHash, bidHash) {
		return fmt.Errorf("hash %x for bid JSON %s does not match hash in shipping: %x",
			calculatedBidJSONHash,
			transientBidJSON,
			bidHash,
		)
	}

	// check 3; check hash of relealed bid matches hash of private bid that was
	// added earlier. This ensures that the bid has not changed since it
	// was added to the shipping

	bidders := shipping.PrivateBids
	privateBidHashString := bidders[bidKey].Hash

	onChainBidHashString := fmt.Sprintf("%x", bidHash)
	if privateBidHashString != onChainBidHashString {
		return fmt.Errorf("hash %s for bid JSON %s does not match hash in shipping: %s, bidder must have changed bid",
			privateBidHashString,
			transientBidJSON,
			onChainBidHashString,
		)
	}

	// we can add the bid to the shipping if all checks have passed
	type transientBidInput struct {
		Price  int    `json:"price"`
		Org    string `json:"org"`
		Bidder string `json:"bidder"`
	}

	// unmarshal bid input
	var bidInput transientBidInput
	err = json.Unmarshal(transientBidJSON, &bidInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// marshal transient parameters and ID and MSPID into bid object
	NewBid := FullBid{
		Type:   bidKeyType,
		Price:  bidInput.Price,
		Org:    bidInput.Org,
		Bidder: bidInput.Bidder,
	}

	// check 4: make sure that the transaction is being submitted is the bidder
	if bidInput.Bidder != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	revealedBids := make(map[string]FullBid)
	revealedBids = shipping.RevealedBids
	revealedBids[bidKey] = NewBid
	shipping.RevealedBids = revealedBids

	newShippingJSON, _ := json.Marshal(shipping)

	// put shipping with bid added back into state
	err = ctx.GetStub().PutState(shippingID, newShippingJSON)
	if err != nil {
		return fmt.Errorf("failed to update shipping: %v", err)
	}

	return nil
}

// CloseShipping can be used by the seller to close the shipping. This prevents
// bids from being added to the shipping, and allows users to reveal their bid
func (s *SmartContract) CloseShipping(ctx contractapi.TransactionContextInterface, shippingID string) error {

	// get shipping from public state
	shipping, err := s.QueryShipping(ctx, shippingID)
	if err != nil {
		return fmt.Errorf("failed to get shipping from public state %v", err)
	}

	// the shipping can only be closed by the seller

	// get ID of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	Seller := shipping.Seller
	if Seller != clientID {
		return fmt.Errorf("shipping can only be closed by seller: %v", err)
	}

	Status := shipping.Status
	if Status != "open" {
		return fmt.Errorf("cannot close shipping that is not open")
	}

	shipping.Status = string("closed")

	closedShippingJSON, _ := json.Marshal(shipping)

	err = ctx.GetStub().PutState(shippingID, closedShippingJSON)
	if err != nil {
		return fmt.Errorf("failed to close shipping: %v", err)
	}

	return nil
}

// EndShipping both changes the shipping status to closed and calculates the winners
// of the shipping
func (s *SmartContract) EndShipping(ctx contractapi.TransactionContextInterface, shippingID string) error {

	// get shipping from public state
	shipping, err := s.QueryShipping(ctx, shippingID)
	if err != nil {
		return fmt.Errorf("failed to get shipping from public state %v", err)
	}

	// Check that the shipping is being ended by the seller

	// get ID of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	Seller := shipping.Seller
	if Seller != clientID {
		return fmt.Errorf("shipping can only be ended by seller: %v", err)
	}

	Status := shipping.Status
	if Status != "closed" {
		return fmt.Errorf("Can only end a closed shipping")
	}

	// get the list of revealed bids
	revealedBidMap := shipping.RevealedBids
	if len(shipping.RevealedBids) == 0 {
		return fmt.Errorf("No bids have been revealed, cannot end shipping: %v", err)
	}

	// determine the highest bid
	for _, bid := range revealedBidMap {
		if bid.Price < shipping.Price {
			shipping.Winner = bid.Bidder
			shipping.Price = bid.Price
		}
	}

	// check if there is a winning bid that has yet to be revealed
	err = checkForHigherBid(ctx, shipping.Price, shipping.RevealedBids, shipping.PrivateBids)
	if err != nil {
		return fmt.Errorf("Cannot end shipping: %v", err)
	}

	shipping.Status = string("ended")

	endedShippingJSON, _ := json.Marshal(shipping)

	err = ctx.GetStub().PutState(shippingID, endedShippingJSON)
	if err != nil {
		return fmt.Errorf("failed to end shipping: %v", err)
	}
	return nil
}
