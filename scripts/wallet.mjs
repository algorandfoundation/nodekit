#!/usr/bin/env node

import {
    Algodv2,
    makeKeyRegistrationTxnWithSuggestedParamsFromObject,
    mnemonicToSecretKey,
    waitForConfirmation
} from 'algosdk'
const key = {
    'mnemonic': 'artefact exist coil life turtle edge edge inside punch glance recycle teach melody diet method pause slam dumb race interest amused side learn able heavy',
    'address': 'TUIDKH2C7MUHZDD77MAMUREJRKNK25SYXB7OAFA6JFBB24PEL5UX4S4GUU',
    'private_key': 'Z/CTWhR4dRnJKHVurdhn6U3F9oRxoVj+0GBbF4Qf20+dEDUfQvsofIx/+wDKRImKmq12WLh+4BQeSUIdceRfaQ=='
}
const account = mnemonicToSecretKey(key.mnemonic)
console.log(account)


const keys = await fetch('http://localhost:8081/v2/participation', {
    headers: {
        "X-Algo-API-Token": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    }
}).then(r => r.json())

const partkey = keys.filter((k)=>k.address === key.address)[0]
console.log(partkey)

const client = new Algodv2(
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "http://localhost",
    8081
)

const params = await client.getTransactionParams().do()


// sets up keys for 100000 rounds
const numRounds = 1e5;

// dilution default is sqrt num rounds
const keyDilution = BigInt(Math.floor(numRounds ** 0.5));

const txn = makeKeyRegistrationTxnWithSuggestedParamsFromObject({
    sender: key.address,
    voteKey: Buffer.from(partkey.key['vote-participation-key'], "base64"),
    selectionKey: Buffer.from(partkey.key['selection-participation-key'], "base64"),
    stateProofKey: Buffer.from(partkey.key['state-proof-key'], "base64"),
    voteFirst: partkey.key['vote-first-valid'],
    voteLast: partkey.key['vote-last-valid'],
    voteKeyDilution: partkey.key['vote-key-dilution'],
    suggestedParams: params,
    }
)

const signtxn = txn.signTxn(account.sk)

const { txId } = await client.sendRawTransaction(signtxn).do();
const result = await waitForConfirmation(client, txId, 40);
console.log(txn)
