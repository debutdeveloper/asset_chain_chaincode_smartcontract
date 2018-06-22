# <b>AssetChain</b>

# Application Architecture
The intent of the AssetChain was to build a limited-use, "production-ready" solution and prove the capability of the blockchain to help with asset lifecycle management issues. Hence, the team picked a set of core user stories that demonstrate business value and overcome technical challenges.

<p align="center">
  <img src="./assetchainArchi1.jpg" width="800"/>
</p>

The solution architecture addresses the following user stories:

- Record an event associated with the asset, such as purchase, inventory, assignment, disposal, and warranty.
- Device request for allocation and related workflow to avoid any discrepancies.
- Determine if device is physically damaged at any location or any event.
- Encrypt and store the original transaction or event details and record the hash of those proofs on the blockchain
- Provide the ability to view the asset lifecycle including the proofs via a user interface using various methods — by asset serial number, and by assignment.
- Data should be accessible only within the organization with private chains

AssetChain is going to manage assets between many users leveraging Hyperledger Fabric. We are going to do this in Node.js and a bit of GoLang. The backend of this application will be the GoLang code running in our blockchain network. From here on out the GoLang code will be referred to as 'chaincode'. The chaincode will be used to create assets, users etc. by storing it to the Blockchain.

## Versions
This version is compatible with Hyperledger Fabric v1.1x.
AssetChain network was successfully tested with the Hyperledger Fabric v1.1.0

## Application Communication flow
<p align="center">
  <img src="./assetchainArchi2.jpg" width="800"/>
</p>

    1. The admin will interact with AssetChain, our Node.js application, in their browser.
    2. This client side JS code will open a websocket to the backend Node.js application. The client JS will send messages to the backend when the admin interacts with the site.
    3. Reading or writing the ledger is known as a proposal. This proposal is built by AssetChain (via the SDK) and then sent to a blockchain peer.
    4. The peer will communicate to its AssetChain chaincode. The chaincode will run/simulate the transaction. If there are no issues it will endorse the transaction and send it back to our AssetChain application.
    5. AssetChain (via the SDK) will then send the endorsed proposal to the ordering service. The orderer will package many proposals from the whole network into a block. Then it will broadcast the new block to peers in the network.
    6. Finally the peer will validate the block and write it to its ledger. The transaction has now taken effect and any subsequent reads will reflect this change.

<hr>

# AssetChain Setup
## 1. Setup Local Environment
First we need to install all dependencies for running this project so we need to download Git, Go, Docker and Node.js. After installing all dependencies go to the next section.
## 2. Download Assetchain

    git clone https://github.com/
## 3. Setup Network

## 4. Setup Assetchain Api

## 5. Setup Admin Pannel

<hr>

# Using AssetChain
1. <b>Admin Panel</b> can be accessed by this url [assetchainadmin.debutinfotech.com](assetchainadmin.debutinfotech.com)
2. <b>Mobile app</b> to access AssetChain can be downloaded from the [Google Play store]() and the [Apple App store]()

<hr>

# BlockChain Background
### Definitions:

- **Peer** - A peer is a member of the blockchain and is running Hyperledger Fabric. From the viewpoint of AssetChain, the peers are owned and operated by [Debut Infotech Pvt. Limited](http://www.debutinfotech.com/)

- **CA**- The CA (Certificate Authority) is responsible for gatekeeping our blockchain network. It will provide transaction certificates for the clients of our AssetChain node.js application.

- **Orderer** - An orderer or ordering service is a member of the blockchain network whose main responsibility is to package transactions into blocks.

- **Users** - A user is an entity that is authorized to interact with the blockchain. In AssetChain, the user can query and write to the ledger.

- **Blocks** - Blocks contain transactions and a hash to verify integrity.

- **Transactions or Proposals** - These represent interactions to the blockchain ledger. A read or write request of the ledger is sent as a transaction/proposal.

- **Ledger** - This is the storage for the blockchain on a peer. It contains the actual block data which consist of transaction parameters and key value pairs. It is written by chaincode.

- **Chaincode** - Chaincode is Hyperledger Fabric speak for smart contracts. It defines the assets and all rules about assets.
<hr>

# AssetChain FAQs
Do you have questions about why something in AssetChain is the way it is? Or how to do something? Check out the [FAQ](./FAQ.md).
<hr>

# Feedback

We are very interested in your feedback. 
If you have any ideas on how to improve the project, please reach out! Specifically:

    -Did the format of the readme work well for you?
    -At which points did you get lost?
    -Is something broken!?
    -Did your knowledge grow by the end of the tutorial?
    -Was something particularly painful?
    -Did it make you have an existential crisis and you -are suddenly unsure of what it means to be, you?

Use the [GitHub Issues]() section to communicate any improvements/bugs and pain points!

