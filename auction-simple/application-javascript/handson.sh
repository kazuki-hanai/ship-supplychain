#!/bin/bash

ss=$(date +%s)

echo '*** 輸送依頼の作成 ***'
node createShip.js 4 org1 seller Ship-$ss コンピュータ 青森 50 5 1>/dev/null
echo ''

echo '*** bidder1 の入札 ***'
node bid.js 4 org1 bidder1 Ship-$ss 500 1>/dev/null 2>bidid
export BIDDER1_BID_ID=$(cat bidid)
node submitBid.js 4 org1 bidder1 Ship-$ss $BIDDER1_BID_ID 1>/dev/null

echo ''
echo '*** bidder2 の入札 ***'
node bid.js 4 org1 bidder2 Ship-$ss 300 1>/dev/null 2>bidid
export BIDDER2_BID_ID=$(cat bidid)
node submitBid.js 4 org1 bidder2 Ship-$ss $BIDDER2_BID_ID 1>/dev/null

echo ''
echo '*** bidder3 の入札 ***'
node bid.js 4 org2 bidder3 Ship-$ss 900 1>/dev/null 2>bidid
export BIDDER3_BID_ID=$(cat bidid)
node submitBid.js 4 org2 bidder3 Ship-$ss $BIDDER3_BID_ID 1>/dev/null

echo ''
echo '*** bidder4 の入札 ***'
node bid.js 4 org2 bidder4 Ship-$ss 800 1>/dev/null 2>bidid
export BIDDER4_BID_ID=$(cat bidid)
node submitBid.js 4 org2 bidder4 Ship-$ss $BIDDER4_BID_ID 1>/dev/null

echo ''
echo '*** オークションのクローズ ***'
node closeShip.js 4 org1 seller Ship-$ss 1>/dev/null

echo ''
echo '*** bidder1 の開示 ***'
node revealBid.js 4 org1 bidder1 Ship-$ss $BIDDER1_BID_ID 1>/dev/null

echo ''
echo '*** bidder2 の開示 ***'
node revealBid.js 4 org1 bidder2 Ship-$ss $BIDDER2_BID_ID 1>/dev/null

echo ''
echo '*** bidder3 の開示 ***'
node revealBid.js 4 org2 bidder3 Ship-$ss $BIDDER3_BID_ID 1>/dev/null

echo ''
echo '*** bidder4 の開示 ***'
node revealBid.js 4 org2 bidder4 Ship-$ss $BIDDER4_BID_ID 1>/dev/null

echo ''
echo '*** オークションの終了 ***'
node endShip.js 4 org1 seller Ship-$ss 1>/dev/null
