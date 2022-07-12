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
const myChaincodeName = 'ship-supplychain_v' + process.argv[2];

async function endShip(ccp,wallet,user,shipID) {
	try {

		const gateway = new Gateway();

		//connect using Discovery enabled
		await gateway.connect(ccp,
			{ wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		// Query the ship to get the list of endorsing orgs.
		let shipString = await contract.evaluateTransaction('QueryShipping',shipID);
		let shipJSON = JSON.parse(shipString);

		let statefulTxn = contract.createTransaction('EndShipping');

		if (shipJSON.organizations.length === 2) {
			statefulTxn.setEndorsingOrganizations(shipJSON.organizations[0],shipJSON.organizations[1]);
		} else {
			statefulTxn.setEndorsingOrganizations(shipJSON.organizations[0]);
		}

		console.log('\n--> Submit the transaction to end the ship');
		await statefulTxn.submit(shipID);
		console.log('*** Result: committed');

		console.log('\n--> Evaluate Transaction: query the updated ship');
		let result = await contract.evaluateTransaction('QueryShipping',shipID);
		console.log('*** Result: Ship: ' + prettyJSONString(result.toString()));
		result = JSON.parse(result.toString());
		console.error('依頼商品: ', result.item.item);
		console.error('輸送先: ', result.item.dest);
		console.error('重さ: ', result.item.org);
		console.error('輸送日数: ', result.item.days);
                console.error('入札(未公開):')
                for (const [key, value] of Object.entries(result.privateBids)) {
                    console.error('    ', value.hash);
                }
                console.error('入札(公開): ');
                for (const [key, value] of Object.entries(result.revealedBids)) {
                    console.error('    ', key, ': ');
                    console.error('        入札価格:', value.price);
                    console.error('        入札者:', value.bidder.slice(9, 16));
                }
		console.error('落札者:', result.winner.slice(9, 16));
		console.error('落札価格:', result.price);
		console.error('オークションステータス:', result.status);

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
			console.log('Usage: node endShip.js org userID shipID');
			process.exit(1);
		}

		const org = process.argv[3];
		const user = process.argv[4];
		const shipID = process.argv[5];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await endShip(ccp,wallet,user,shipID);
		}
		else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await endShip(ccp,wallet,user,shipID);
		}  else {
			console.log('Usage: node endShip.js org userID shipID');
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
