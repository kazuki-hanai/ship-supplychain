/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const {
	buildCCPOrg1,
	buildCCPOrg2,
	buildWallet,
	prettyJSONString,
} = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'ship-supplychain_v' + process.argv[2];

async function createShip(ccp, wallet, user, auctionID, itemName, itemDest, itemWeight, itemDays) {
	try {
		const gateway = new Gateway();

		//connect using Discovery enabled
		await gateway.connect(ccp, {
			wallet: wallet,
			identity: user,
			discovery: { enabled: true, asLocalhost: true },
		});

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		let statefulTxn = contract.createTransaction('CreateShipping');

		console.log('\n--> Submit Transaction: Propose a new auction');
		await statefulTxn.submit(auctionID, itemName, itemDest, itemWeight, itemDays);
		console.log('*** Result: committed');

		console.log(
			'\n--> Evaluate Transaction: query the auction that was just created'
		);
		let result = await contract.evaluateTransaction('QueryShipping', auctionID);
		console.log('*** Result: Ship: ' + prettyJSONString(result.toString()));

		gateway.disconnect();
	} catch (error) {
		console.error(`******** FAILED to submit bid: ${error}`);
	}
}

async function main() {
	try {
		if (
			process.argv[2] === undefined ||
      process.argv[3] === undefined ||
      process.argv[4] === undefined ||
      process.argv[5] === undefined ||
      process.argv[6] === undefined ||
      process.argv[7] === undefined ||
      process.argv[8] === undefined ||
      process.argv[9] === undefined
		) {
			console.log('Usage: node createShip.js contractVersion org userID auctionID itemName itemDest itemWeight');
			process.exit(1);
		}

		const org = process.argv[3];
		const user = process.argv[4];
		const auctionID = process.argv[5];
		const itemName = process.argv[6];
		const itemDest = process.argv[7];
		const itemWeight = process.argv[8];
		const itemDays = process.argv[9];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await createShip(ccp, wallet, user, auctionID, itemName, itemDest, itemWeight, itemDays);
		} else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await createShip(ccp, wallet, user, auctionID, itemName, itemDest, itemWeight, itemDays);
		} else {
			console.log('Usage: node createShip.js org userID auctionID item');
			console.log('Org must be Org1 or Org2');
		}
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
	}
}

main();
