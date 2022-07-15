## 物流システムの最適化

ブラインドオークションとブロックチェーンの仕組みを利用して物流のシステムを最適化します．ブラインドオークションとは，オークションが終わるまで，入札価格が他者に見えないようなオークションです．配送元は，配送する商品の名前や価格，配送日数を入力します．そして，輸送会社がそれぞれ，価格を入札します．オークションが終わった時に最も安い値段を付けた会社がその商品の配送権を得ます．このシステムを利用することで，配送元はより安い価格で商品の輸送を行うことができ，輸送会社は幅広い商品輸送の依頼を受けることができるようになります．これらのロジックはすべてブロックチェーン上に実装されているのでこのシステムを利用することで他社の信頼にかかわらず，サプライチェーンのなかでも，流通に関わる部分の最適化を行うことができます．

## ネットワークの作成
```
cd ./test-network
./network.sh up
```

## チャネルの作成
```
./network.sh createChannel
```

## コードをデプロイ

```
./network.sh deployCC -ccn ship-supplychain_v1 -ccp ../ship-supplychain/chaincode-go  -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')"
```

## アプリケーションのインストール

```
cd application-javascript
npm install
```

## Adminとユーザの登録

```
node enrollAdmin.js org1
node enrollAdmin.js org2
```

```
node registerEnrollUser.js org1 seller
node registerEnrollUser.js org1 bidder1
node registerEnrollUser.js org1 bidder2
node registerEnrollUser.js org2 bidder3
node registerEnrollUser.js org2 bidder4
```

## 輸送依頼の作成

```
node createShip.js 1 org1 seller ComputerShip コンピュータ 青森県 50 5
```

## 入札

### 入札者1
入札情報の作成
```
node bid.js 1 org1 bidder1 ComputerShip 800
```

入札情報の送信
```
export BIDDER1_BID_ID=<BID_ID>
node submitBid.js 1 org1 bidder1 ComputerShip $BIDDER1_BID_ID
```


### 入札者２
入札情報の作成
```
node bid.js 1 org1 bidder2 ComputerShip 500
```

入札情報の送信
```
export BIDDER2_BID_ID=<BID_ID>
node submitBid.js 1 org1 bidder2 ComputerShip $BIDDER2_BID_ID
```

### 入札者３
入札情報の作成
```
node bid.js 1 org2 bidder3 ComputerShip 700
```

入札情報の送信
```
export BIDDER3_BID_ID=<BID_ID>
node submitBid.js 1 org2 bidder3 ComputerShip $BIDDER3_BID_ID
```

### 入札者４
入札情報の作成
```
node bid.js 1 org2 bidder4 ComputerShip 900
```

入札情報の送信
```
export BIDDER4_BID_ID=<BID_ID>
node submitBid.js 1 org2 bidder4 ComputerShip $BIDDER4_BID_ID
```

## 入札の終了

```
node closeShip.js 1 org1 seller ComputerShip 
```

## 入札価格の開示

```
node revealBid.js 1 org1 bidder1 ComputerShip $BIDDER1_BID_ID
```

```
node revealBid.js 1 org2 bidder3 ComputerShip $BIDDER3_BID_ID
```

```
node revealBid.js 1 org2 bidder4 ComputerShip $BIDDER4_BID_ID
```

## オークションの終了

```
node endShip.js 1 org1 seller ComputerShip
```
