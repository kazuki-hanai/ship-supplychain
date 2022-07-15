/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet } = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'ship-supplychain_v' + process.argv[2];


function prettyJSONString(inputString) {
	if (inputString) {
		return JSON.stringify(JSON.parse(inputString), null, 2);
	}
	else {
		return inputString;
	}
}

async function submitBid(ccp,wallet,user,shipID,bidID) {
	try {

		const gateway = new Gateway();

		//connect using Discovery enabled
		await gateway.connect(ccp,
			{ wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		// console.log('\n--> Evaluate Transaction: query the ship you want to join');
		let shipString = await contract.evaluateTransaction('QueryShipping',shipID);
		let shipJSON = JSON.parse(shipString);

		let statefulTxn = contract.createTransaction('SubmitBid');

		if (shipJSON.organizations.length === 2) {
			statefulTxn.setEndorsingOrganizations(shipJSON.organizations[0],shipJSON.organizations[1]);
		} else {
			statefulTxn.setEndorsingOrganizations(shipJSON.organizations[0]);
		}

		// console.log('\n--> Submit Transaction: add bid to the ship');
		await statefulTxn.submit(shipID,bidID);

		// console.log('\n--> Evaluate Transaction: query the ship to see that our bid was added');
		let result = await contract.evaluateTransaction('QueryBid', shipID, bidID);
		result = JSON.parse(result.toString());
                console.error('入札者: ', result.bidder.slice(9, 16));
                console.error('価格: ', result.price);

		gateway.disconnect();
	} catch (error) {
		console.error(`******** FAILED to submit bid: ${error}`);
		process.exit(1);
	}
}

async function main() {
	try {

		if (process.argv[2] === undefined || process.argv[3] === undefined ||
            process.argv[4] === undefined || process.argv[5] === undefined || process.argv[6] === undefined) {
			console.log('Usage: node submitBid.js org userID shipID bidID');
			process.exit(1);
		}

		const org = process.argv[3];
		const user = process.argv[4];
		const shipID = process.argv[5];
		const bidID = process.argv[6];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await submitBid(ccp,wallet,user,shipID,bidID);
		}
		else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await submitBid(ccp,wallet,user,shipID,bidID);
		}
		else {
			console.log('Usage: node submitBid.js org userID shipID bidID');
			console.log('Org must be Org1 or Org2');
		}
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
		if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}
}


main();
