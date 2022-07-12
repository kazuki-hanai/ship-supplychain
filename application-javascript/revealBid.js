/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const { buildCCPOrg1, buildCCPOrg2, buildWallet, prettyJSONString} = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'ship-supplychain_v2';

async function addBid(ccp,wallet,user,shipID,bidID) {
	try {

		const gateway = new Gateway();
		await gateway.connect(ccp,
			{ wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		console.log('\n--> Evaluate Transaction: read your bid');
		let bidString = await contract.evaluateTransaction('QueryBid',shipID,bidID);
		let bidJSON = JSON.parse(bidString);

		//console.log('\n--> Evaluate Transaction: query the ship you want to join');
		let shipString = await contract.evaluateTransaction('QueryShipping',shipID);
		// console.log('*** Result:  Bid: ' + prettyJSONString(shipString.toString()));
		let shipJSON = JSON.parse(shipString);

		let bidData = { objectType: 'bid', price: parseInt(bidJSON.price), org: bidJSON.org, bidder: bidJSON.bidder};
		console.log('*** Result:  Bid: ' + JSON.stringify(bidData,null,2));

		let statefulTxn = contract.createTransaction('RevealBid');
		let tmapData = Buffer.from(JSON.stringify(bidData));
		statefulTxn.setTransient({
			bid: tmapData
		});

		if (shipJSON.organizations.length === 2) {
			statefulTxn.setEndorsingOrganizations(shipJSON.organizations[0],shipJSON.organizations[1]);
		} else {
			statefulTxn.setEndorsingOrganizations(shipJSON.organizations[0]);
		}

		await statefulTxn.submit(shipID,bidID);

		console.log('\n--> Evaluate Transaction: query the ship to see that our bid was added');
		let result = await contract.evaluateTransaction('QueryShipping',shipID);
		console.log('*** Result: Ship: ' + prettyJSONString(result.toString()));

		gateway.disconnect();
	} catch (error) {
		console.error(`******** FAILED to submit bid: ${error}`);
		process.exit(1);
	}
}

async function main() {
	try {

		if (process.argv[2] === undefined || process.argv[3] === undefined ||
            process.argv[4] === undefined || process.argv[5] === undefined) {
			console.log('Usage: node revealBid.js org userID shipID bidID');
			process.exit(1);
		}

		const org = process.argv[2];
		const user = process.argv[3];
		const shipID = process.argv[4];
		const bidID = process.argv[5];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await addBid(ccp,wallet,user,shipID,bidID);
		}
		else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await addBid(ccp,wallet,user,shipID,bidID);
		}
		else {
			console.log('Usage: node revealBid.js org userID shipID bidID');
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
