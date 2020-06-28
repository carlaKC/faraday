# Accounting Reports
Faraday produces accounting reports on a node's on chain and off chain activity. 
These reports are formatted using the [Harmony Reporting Standard](https://github.com/picksco/harmony). 
This document provides a description of the entries in these reports. 

## Common Fields
For brevity, the following fields which have the same meaning for each entry will be omitted: 
- Timestamp: The timestamp of the block that the channel open transaction appeared in. 
- Fiat: The value of the amount field in USD. Note that values less than one satoshi will be rounded down to zero. 
- OnChain: Whether the transaction occurred off chain, or on chain.
- Credit: True when an entry increased our balances, false when an entry decreased our balances. 

Note that fee entries reference the entry they are associated with by appending a fee marker (:-1) to the original reference. The fee entry will have a reference formatted as follows: `original reference:-1`. 

## On Chain Reports

### Local Channel Open
Local channel open entry types represent channel opens that were initiated by 
our node. These entries are accompanied by a separate Channel Open Fees entry, 
because the opening party pays on chain fees. 

- Amount: The amount in millisatoshis that we added to the channel, excluding on chain fees. 
- TxID: The on chain transaction ID for the channel open. 
- Reference: The unique channel ID assigned to the channel. 
- Note: A note with details of who opened the channel. 

### Channel Open Fees
The fees paid to open a channel that we initiated. 

- Amount: The amount in millisatoshis of on chain fees paid.  
- TxID: The on chain transaction ID for the channel open. 
- Reference: TransactionID:-1; note that this is a special marker for fees. 
- Note: A note containing the pubkey of the peer that we opened the channel to. 

### Remote Channel Open
Remote channel open entry types represent channels that were opened by remote
peers. 

- Amount: Zero, our balance is unaffected by remote channel creation, with the exception of a push amount listed below. 
- TxID: The on chain transaction ID for the channel open. 
- Reference: The unique channel ID assigned to the channel. 
- Note: A note containing the pubkey of the peer that opened a channel to us. 

Known Omissions:
- Remote peers may push balance to our node as part of the funding flow. This amount is not currently included in these reports. 

### Channel Close 
Channel close entries represent the on chain close of a channel. 

- Amount: The amount in millisatoshis that was paid out to us immediately on channel close. 
- TxID: The on chain transaction ID for the channel close. 
- Reference: The channel close transaction ID.
- Note: A note indicating the type of channel close, and who initiated it. 

Known Omissions: 
- If our balance is encumbered behind a timelock, or in an unresolved htlc, it will not be paid out as part of this transaction and must be resolved by follow up on chain transactions. 
- The fees paid to close channels that we initiated are not currently recorded, this is because balances are taken from the funding output rather than being supplied by the wallet.

### Receipt
A receipt is an on chain transaction which paid to our wallet which was not related to the opening/closing of channels.

- Amount: The amount in millisatoshis that was paid to an address controlled by our wallet.
- TxID: The on chain transaction ID.
- Reference: The on chain transaction ID.
- Note: An optional label set on transaction publish (see [lnd transaction labels](https://github.com/lightningnetwork/lnd/blob/master/lnrpc/walletrpc/walletkit.proto#L136)). 

Known Omissions:
- This entry type will include on chain resolutions for channel closes that sweep balances back to our node.

### Payment
A payment is an on chain transaction which was paid from our wallet and was not related to the opening/closing of channels. 
- Amount: The amount in millisatoshis that was paid from an address controlled by our wallet.
- TxID: The on chain transaction ID.
- Reference: The on chain transaction ID.
- Note: An optional label set on transaction publish (see [lnd transaction labels](https://github.com/lightningnetwork/lnd/blob/master/lnrpc/walletrpc/walletkit.proto#L136)). 

Known Omissions:
- This entry type will include the on chain resolution of htlcs when we force close on our peers and have to settle or fail them on chain. 
- The current accounting package does not support accounting for payments with duplicate payment hashes, which were allowed in previous versions of lnd. Duplicate payments should be deleted or a time range that does not include them should be specified. 

### Fee
A fee entry represents the on chain fees we paid for a transaction. 

- Amount: The amount in millisatoshis that was paid in fees from our wallet. 
- TxID: The on chain transaction ID.
- Reference: TransactionID:-1. 
- Note: Note set for fees. 

### Sweep
A sweep is an on chain transaction which is used to sweep our own funds back to our wallet. This is required when a channel is force closed and need to sweep time locked commitment outputs, htlcs or both. 

- TxID: The on chain transaction ID.
- Reference: The on chain transaction ID.
- Note: An optional label set on transaction publish (see [lnd transaction labels](https://github.com/lightningnetwork/lnd/blob/master/lnrpc/walletrpc/walletkit.proto#L136)). 

Known Omissions:
- Note that this entry type does not include the first stage success/timeout transactions that are required to resolve htlcs when we force close on our peer. We currently do not have the information required to identify these transactions available, so they are included in payments, see known omissions. 

### Sweep Fee
A fee entry represents the on chain fees we paid for a sweep.

- Amount: The amount in millisatoshis that was paid in fees from our wallet. 
- TxID: The on chain transaction ID.
- Reference: TransactionID:-1. 
- Note: Not set for fees. 

Known Omissions: 
- If the wallet did not have knowledge that an input was owned by lnd (which is the case for more complex scripts, like htlcs), then it will not record fees that are siphoned off the input amount here because it does not know that we control those funds. 

### Swap Success
A swap success transaction is one that spends from an on chain htlc offered as part of a sub-atomic swap using the preimage. When completing a successful [loop out swap](https://lightning.engineering/posts/2019-04-15-loop-out-in-depth/), this transaction sweeps funds to your wallet. 

- Amount: The amount in millisatoshis that was paid to an address controlled by our wallet.
- TxID: The on chain transaction ID.
- Reference: The on chain transaction ID.
- Note: An optional label set on transaction publish (see [lnd transaction labels](https://github.com/lightningnetwork/lnd/blob/master/lnrpc/walletrpc/walletkit.proto#L136)). 

Known Omissions:
- We will only identify swap success transactions if they follow the format used by [Lightning Labs Loop](https://github.com/lightninglabs/loop/blob/master/swap/htlc.go#L147). 

### Swap Success Fee
The fees paid for the transaction above. 

- Amount: The amount in millisatoshis that was paid in fees from our wallet. 
- TxID: The on chain transaction ID.
- Reference: TransactionID:-1. 
- Note: Not set for fees. 

### Swap Timeout
A swap timeout transaction is one that spends from an on chain htlc offered as part of a sub-atomic swap using the timeout clause. If a [loop in swap](https://blog.lightning.engineering/announcement/2019/06/25/loop-in.html) fails, this transaction sweeps funds to your wallet. 

- Amount: The amount in millisatoshis that was paid to an address controlled by our wallet.
- TxID: The on chain transaction ID.
- Reference: The on chain transaction ID.
- Note: An optional label set on transaction publish (see [lnd transaction labels](https://github.com/lightningnetwork/lnd/blob/master/lnrpc/walletrpc/walletkit.proto#L136)). 

Known Omissions:
- We will only identify swap success transactions if they follow the format used by [Lightning Labs Loop](https://github.com/lightninglabs/loop/blob/master/swap/htlc.go#L147). 

### Swap Timeout Fee
The fees paid for the transaction above. 

- Amount: The amount in millisatoshis that was paid in fees from our wallet. 
- TxID: The on chain transaction ID.
- Reference: TransactionID:-1. 
- Note: Not set for fees. 

## Off Chain Reports

### Receipt
Receipts off chain represent invoices that are paid via the Lightning Network.

- Amount: The amount in millisatoshis that we were paid, note that this may be greater than the original invoice value.
- TxID: The payment hash of the invoice.
- Reference: The preimage of the invoice.
- Note: Optionally set if the invoice had a memo attached, was overpaid, or was a keysend.

### Circular Receipt
Circular receipts record instances where we have paid one of our own invoices. 

- Amount: The amount in millisatoshis that we were paid, note that this may be greater than the original invoice value.
- TxID: The payment hash of the invoice.
- Reference: The preimage of the invoice.
- Note: Optionally set if the invoice had a memo attached, was overpaid, or was a keysend.

### Payment
Payments off chain represent payments made via the Lightning Network. 

- Amount: The amount in millisatoshis that we paid, excluding the off chain fees paid. 
- TxID: The payment hash. 
- Reference: Unique payment ID: Payment hash. 
- Note: The preimage for the payment, which serves as proof of payment.

### Fee
- Amount: The amount in millisatoshis that was paid in off chain fees. 
- TxID: The payment hash. 
- Reference: Unique payment ID: Payment hash: -1. 
- Note: A note indicating the number of htlcs the payment was paid over.

### Circular Payment
Circular payments represent payments made to our own node to rebalance channels. These payments are paid from our node to one of our own invoices.

- Amount: The amount that was rebalanced.
- TxID: The payment hash.
- Reference: Unique payment ID: Payment hash. 
- Note: The preimage for the payment, which serves as proof of payment.

### Circular Payment Fee
Circular payment fees represent the fees we paid to loop a circular payment to ourselves.

- Amount: The amount that was paid in off chain fees. 
- TxID: The payment hash.
- Reference: Unique payment ID: Payment hash: -1. 
- Note: A note indicating the number of htlcs the payment was paid over.

### Forwards
A forward represents a payment that arrives at our node on an incoming channel and is forwarded out on an outgoing channel in exchange for fees. The forward itself does not changes our balance, since it just shifts funds over our channels. We include forwarding entries with zero balances for completeness. Forwarding fee entries reflect the increase in our holdings from the fee we are paid. 

- Amount: Zero, forwards do not change our balance except for fees, which are separated out.
- TxID:  Timestamp: Incoming Channel ID: Outgoing Channel ID.
- Reference: Not set for forwards.
- Note: The amounts that were forwarded in and out of our node. 

Known Omissions: 
- We use incoming and outgoing channel ID paired with timestamp as a best-effort version of a txid. Note that this is not strictly unique for a single htlc, it is theoretically possible for two htlcs to pass through the same channel with the same timestamp.

### Forward Fee
Forward fee entries represent the fees we earned from forwarding payments. 

- Amount: The amount in millisatoshis of fees we earned from the forward. 
- TxID: Timestamp: Incoming Channel ID: Outgoing Channel ID.
- Reference: Not set for forwards.
- Note: Not set for forwards.

Known Omissions: 
- See the note on txids in the Forwards section. 

