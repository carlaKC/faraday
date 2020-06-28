package accounting

import (
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/lightninglabs/lndclient"
	"github.com/lightninglabs/loop/swap"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/routing/route"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
	testKey = [33]byte{
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1,
	}

	params = &chaincfg.TestNet3Params

	openChannelTx              = "44183bc482d5b7be031739ce39b6c91562edd882ba5a9e3647341262328a2228"
	remotePubkey               = "02f6a7664ca2a2178b422a058af651075de2e5bdfff028ac8e1fcd96153cba636b"
	remoteVertex, _            = route.NewVertexFromStr(remotePubkey)
	channelID           uint64 = 124244814004224
	channelCapacitySats        = btcutil.Amount(500000)
	channelFeesSats            = btcutil.Amount(10000)

	openChannel = lndclient.ChannelInfo{
		PubKeyBytes:  remoteVertex,
		ChannelPoint: fmt.Sprintf("%v:%v", openChannelTx, 1),
		ChannelID:    channelID,
		Capacity:     channelCapacitySats,
		Initiator:    false,
	}

	transactionTimestamp = time.Unix(1588145604, 0)

	openChannelTransaction = lndclient.Transaction{
		TxHash: openChannelTx,
		// Amounts are reported with negative values in getTransactions.
		Amount:        channelCapacitySats * -1,
		Confirmations: 2,
		Fee:           channelFeesSats,
		Timestamp:     transactionTimestamp,
	}

	mockPrice = decimal.NewFromInt(10)

	closeTx = "e730b07d6121b19dd717925de82b8c76dec38517ffd85701e6735a726f5f75c3"

	closeBalanceSat = btcutil.Amount(50000)

	channelClose = lndclient.ClosedChannel{
		ChannelPoint:   openChannel.ChannelPoint,
		ChannelID:      openChannel.ChannelID,
		ClosingTxHash:  closeTx,
		PubKeyBytes:    remoteVertex,
		SettledBalance: closeBalanceSat,
		CloseInitiator: lndclient.InitiatorLocal,
	}

	closeTimestamp = time.Unix(1588159722, 0)

	channelCloseTx = lndclient.Transaction{
		TxHash:    closeTx,
		Amount:    closeBalanceSat,
		Timestamp: closeTimestamp,
		// Total fees for closes will always reflect as 0 because they
		// come from the 2-2 multisig funding output.
		Fee: 0,
	}

	onChainTxID      = "e75760156b04234535e6170f152697de28b73917c69dda53c60baabdae571457"
	onChainAmtSat    = btcutil.Amount(10000)
	onChainFeeSat    = btcutil.Amount(1000)
	onChainTimestamp = time.Unix(1588160816, 0)

	currentHeight   uint32 = 600000
	txConfirmations int32  = 9

	// multiInputTx is a transaction which does not match our template for
	// loop out and loop in transactions (it has too many inputs).
	multiInputTx = &wire.MsgTx{
		TxIn: []*wire.TxIn{{}, {}},
	}

	paymentRequest = "lnbcrt10n1p0t6nmypp547evsfyrakg0nmyw59ud9cegkt99yccn5nnp4suq3ac4qyzzgevsdqqcqzpgsp54hvffpajcyddm20k3ptu53930425hpnv8m06nh5jrd6qhq53anrq9qy9qsqphhzyenspf7kfwvm3wyu04fa8cjkmvndyexlnrmh52huwa4tntppjmak703gfln76rvswmsx2cz3utsypzfx40dltesy8nj64ttgemgqtwfnj9"

	invoiceMemo = "memo"

	invoiceAmt = lnwire.MilliSatoshi(300)

	invoiceOverpaidAmt = lnwire.MilliSatoshi(400)

	invoiceSettleTime = time.Unix(1588159722, 0)

	invoicePreimage = "b5f0c5ac0c873a05702d0aa63a518ecdb8f3ba786be2c4f64a5b10581da976ae"
	preimage, _     = lntypes.MakePreimageFromStr(invoicePreimage)

	invoiceHash = "afb2c82483ed90f9ec8ea178d2e328b2ca526313a4e61ac3808f715010424659"
	hash, _     = lntypes.MakeHashFromStr(invoiceHash)

	invoice = lndclient.Invoice{
		Memo:           invoiceMemo,
		Preimage:       &preimage,
		Hash:           hash,
		Amount:         invoiceAmt,
		SettleDate:     invoiceSettleTime,
		PaymentRequest: paymentRequest,
		AmountPaid:     invoiceOverpaidAmt,
		IsKeysend:      true,
	}

	paymentTime = time.Unix(1590399649, 0)

	paymentHash = "11f414479f0a0c2762492c71c58dded5dce99d56d65c3fa523f73513605bebb3"
	pmtHash, _  = lntypes.MakeHashFromStr(paymentHash)

	paymentPreimage = "adfef20b24152accd4ed9a05257fb77203d90a8bbbe6d4069a75c5320f0538d9"
	pmtPreimage, _  = lntypes.MakePreimageFromStr(paymentPreimage)

	paymentMsat = 30000

	paymentFeeMsat = 45

	paymentIndex = 33

	payment = lndclient.Payment{
		Hash:     pmtHash,
		Preimage: &pmtPreimage,
		Amount:   lnwire.MilliSatoshi(paymentMsat),
		Status: &lndclient.PaymentStatus{
			State: lnrpc.Payment_SUCCEEDED,
		},
		Fee:            lnwire.MilliSatoshi(paymentFeeMsat),
		Htlcs:          []*lnrpc.HTLCAttempt{{}},
		SequenceNumber: uint64(paymentIndex),
	}

	settledPmt = settledPayment{
		Payment:    payment,
		settleTime: paymentTime,
	}

	forwardTs = time.Unix(1590578022, 0)

	forwardChanIn  uint64 = 130841883770880
	forwardChanOut uint64 = 124244814004224

	fwdInMsat  = lnwire.MilliSatoshi(4000)
	fwdOutMsat = lnwire.MilliSatoshi(3000)
	fwdFeeMsat = lnwire.MilliSatoshi(1000)

	fwdEntry = lndclient.ForwardingEvent{
		Timestamp:     forwardTs,
		ChannelIn:     forwardChanIn,
		ChannelOut:    forwardChanOut,
		FeeMsat:       fwdFeeMsat,
		AmountMsatIn:  fwdInMsat,
		AmountMsatOut: fwdOutMsat,
	}
)

// mockConvert is a mocked price function which returns mockPrice * amount.
func mockConvert(amt int64, _ time.Time) (decimal.Decimal, error) {
	amtDecimal := decimal.NewFromInt(amt)
	return mockPrice.Mul(amtDecimal), nil
}

// TestChannelOpenEntry tests creation of entries for locally and remotely
// initiated channels. It uses the globally declared open channel and tx to
// as input.
func TestChannelOpenEntry(t *testing.T) {
	// Returns a channel entry for the constant channel.
	getChannelEntry := func(initiator bool) *HarmonyEntry {
		var (
			amt       = satsToMsat(channelCapacitySats)
			credit    = false
			entryType = EntryTypeLocalChannelOpen
		)

		if !initiator {
			amt = 0
			credit = true
			entryType = EntryTypeRemoteChannelOpen
		}

		mockFiat, _ := mockConvert(amt, transactionTimestamp)

		note := channelOpenNote(
			initiator, remotePubkey, channelCapacitySats,
		)

		return &HarmonyEntry{
			Timestamp: transactionTimestamp,
			Amount:    lnwire.MilliSatoshi(amt),
			FiatValue: mockFiat,
			TxID:      openChannelTx,
			Reference: fmt.Sprintf("%v", channelID),
			Note:      note,
			Type:      entryType,
			OnChain:   true,
			Credit:    credit,
		}

	}

	feeAmt := satsToMsat(channelFeesSats)
	mockFee, _ := mockConvert(feeAmt, transactionTimestamp)
	// Fee entry is the expected fee entry for locally initiated channels.
	feeEntry := &HarmonyEntry{
		Timestamp: transactionTimestamp,
		Amount:    lnwire.MilliSatoshi(feeAmt),
		FiatValue: mockFee,
		TxID:      openChannelTx,
		Reference: feeReference(openChannelTx),
		Note:      channelOpenFeeNote(channelID),
		Type:      EntryTypeChannelOpenFee,
		OnChain:   true,
		Credit:    false,
	}

	tests := []struct {
		name string

		// initiator is used to set the initiator bool on the open
		// channel struct we use.
		initiator bool

		expectedErr error
	}{
		{
			name:      "remote initiator",
			initiator: false,
		},
		{
			name:      "local initiator",
			initiator: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			// Make local copies of the global vars so that we can
			// safely change fields.
			channel := openChannel
			tx := openChannelTransaction

			// Set the initiator field according to test
			// requirement.
			channel.Initiator = test.initiator

			// Get our entries.
			entries, err := channelOpenEntries(
				channel, tx, mockConvert,
			)
			require.Equal(t, test.expectedErr, err)

			// At a minimum, we expect a channel entry to be present.
			expectedChanEntry := getChannelEntry(test.initiator)

			// If we opened the chanel, we also expect a fee entry
			// to be present.
			expected := []*HarmonyEntry{expectedChanEntry}
			if test.initiator {
				expected = append(expected, feeEntry)
			}

			require.Equal(t, expected, entries)
		})
	}
}

// TestChannelCloseEntry tests creation of channel close entries.
func TestChannelCloseEntry(t *testing.T) {
	// getCloseEntry returns a close entry for the global close var with
	// correct close type and amount.
	getCloseEntry := func(closeType, closeInitiator string,
		closeBalance btcutil.Amount) *HarmonyEntry {

		note := channelCloseNote(channelID, closeType, closeInitiator)

		closeAmt := satsToMsat(closeBalance)
		closeFiat, _ := mockConvert(closeAmt, closeTimestamp)

		return &HarmonyEntry{
			Timestamp: closeTimestamp,
			Amount:    lnwire.MilliSatoshi(closeAmt),
			FiatValue: closeFiat,
			TxID:      closeTx,
			Reference: closeTx,
			Note:      note,
			Type:      EntryTypeChannelClose,
			OnChain:   true,
			Credit:    true,
		}
	}

	tests := []struct {
		name      string
		closeType lndclient.CloseType
		closeAmt  btcutil.Amount
	}{
		{
			name:      "coop close, has balance",
			closeType: lndclient.CloseTypeCooperative,
			closeAmt:  closeBalanceSat,
		},
		{
			name:      "force close, has no balance",
			closeType: lndclient.CloseTypeLocalForce,
			closeAmt:  0,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			// Make copies of the global vars so we can change some
			// fields.
			closeChan := channelClose
			closeTx := channelCloseTx

			closeChan.CloseType = test.closeType
			closeTx.Amount = test.closeAmt

			entries, err := closedChannelEntries(
				closeChan, closeTx, mockConvert,
			)
			require.NoError(t, err)

			expected := []*HarmonyEntry{getCloseEntry(
				test.closeType.String(),
				closeChan.CloseInitiator.String(),
				test.closeAmt,
			)}

			require.Equal(t, expected, entries)
		})
	}
}

// TestOnChainEntry tests creation of entries for receipts and payments, and the
// generation of a fee entry where applicable.
func TestOnChainEntry(t *testing.T) {
	// Create a valid htlc which we can use to match entries with sweep
	// spends.
	validNP2WSH, err := swap.NewHtlc(
		int32(currentHeight), testKey, testKey, hash, swap.HtlcNP2WSH,
		params,
	)
	require.NoError(t, err)

	timeOut := []*wire.TxIn{
		{
			Witness: [][]byte{
				{},
				{0},
				validNP2WSH.Script,
			},
		},
	}

	success := []*wire.TxIn{
		{
			Witness: [][]byte{
				{},
				preimage[:],
				validNP2WSH.Script,
			},
		},
	}

	tests := []struct {
		name string

		// Amount is the amount of our on chain tx in sats.
		amount btcutil.Amount

		// Fee amount is the amount of our fee in sats. If the amount
		// is zero, the test does not expect a fee entry to be produced.
		feeAmount btcutil.Amount

		// isSweep indicates whether the transaction should be flagged
		// as a sweep.
		isSweep bool

		// txLabel is an optional label on the rpc transaction.
		txLabel string

		// tx contains the raw transaction we want to test with, we
		// allow this to vary so that we can identify swap success and
		// failure entries.
		tx *wire.MsgTx

		// The entry type we expect to be made for this test.
		entryType EntryType

		// The fee type we expect to be made for this test, if any.
		feeType EntryType
	}{
		{
			name:      "receive with fee",
			amount:    onChainAmtSat,
			feeAmount: onChainFeeSat,
			isSweep:   false,
			tx:        multiInputTx,
			entryType: EntryTypeReceipt,
			feeType:   EntryTypeFee,
		},
		{
			name:      "receive without fee",
			amount:    onChainAmtSat,
			feeAmount: 0,
			isSweep:   false,
			tx:        multiInputTx,
			entryType: EntryTypeReceipt,
		},
		{
			name:      "payment without fee",
			amount:    onChainAmtSat * -1,
			feeAmount: 0,
			isSweep:   false,
			tx:        multiInputTx,
			entryType: EntryTypePayment,
		},
		{
			name:      "payment with fee",
			amount:    onChainAmtSat * -1,
			feeAmount: onChainFeeSat,
			isSweep:   false,
			tx:        multiInputTx,
			entryType: EntryTypePayment,
			feeType:   EntryTypeFee,
		},
		{
			name:      "sweep with fee",
			amount:    onChainAmtSat,
			feeAmount: onChainFeeSat,
			isSweep:   true,
			tx:        multiInputTx,
			entryType: EntryTypeSweep,
			feeType:   EntryTypeSweepFee,
		},
		{
			name:    "success tx",
			amount:  onChainAmtSat,
			isSweep: false,
			tx: &wire.MsgTx{
				TxIn: success,
			},
			entryType: EntryTypeSwapSuccess,
		},
		{
			name:    "timeout tx",
			amount:  onChainAmtSat,
			isSweep: false,
			tx: &wire.MsgTx{
				TxIn: timeOut,
			},
			entryType: EntryTypeSwapTimeout,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			// Create a basic on chain transaction we will use
			// for testing, using the amount, fee amount and raw
			// transaction and label provided by the test.
			onChainTx := lndclient.Transaction{
				TxHash:        onChainTxID,
				Amount:        test.amount,
				Fee:           test.feeAmount,
				Timestamp:     onChainTimestamp,
				Confirmations: txConfirmations,
				Tx:            test.tx,
				Label:         test.txLabel,
			}

			entries, err := onChainEntries(
				onChainTx, test.isSweep, currentHeight,
				mockConvert,
			)
			require.NoError(t, err)

			// Create the entry we expect. We report the absolute
			// amount for transactions with a credit field to
			// indicate whether our amount is positive or negative
			// so we get the absolute amount here.
			amt := satsToMsat(test.amount)
			if test.amount < 0 {
				amt *= -1
			}

			fiat, _ := mockConvert(amt, onChainTimestamp)

			expected := []*HarmonyEntry{
				{
					Timestamp: onChainTimestamp,
					Amount:    lnwire.MilliSatoshi(amt),
					FiatValue: fiat,
					TxID:      onChainTxID,
					Reference: onChainTxID,
					Note:      test.txLabel,
					Type:      test.entryType,
					OnChain:   true,
					Credit:    test.amount >= 0,
				},
			}

			// If we have a fee amount, add it to our set of
			// expected entries.
			if test.feeAmount != 0 {
				feeAmt := satsToMsat(test.feeAmount)
				fiat, _ = mockConvert(feeAmt, onChainTimestamp)

				feeEntry := &HarmonyEntry{
					Timestamp: onChainTimestamp,
					Amount:    lnwire.MilliSatoshi(feeAmt),
					FiatValue: fiat,
					TxID:      onChainTxID,
					Reference: feeReference(onChainTxID),
					Note:      "",
					Type:      test.feeType,
					OnChain:   true,
					Credit:    false,
				}

				expected = append(expected, feeEntry)
			}

			require.Equal(t, expected, entries)
		})
	}
}

// TestInvoiceEntry tests creation of entries for regular invoices and circular
// receipts.
func TestInvoiceEntry(t *testing.T) {
	getEntry := func(circular bool) *HarmonyEntry {
		note := invoiceNote(
			invoice.Memo, invoice.Amount, invoice.AmountPaid,
			invoice.IsKeysend,
		)

		fiat, _ := mockConvert(
			int64(invoiceOverpaidAmt), invoiceSettleTime,
		)

		expectedEntry := &HarmonyEntry{
			Timestamp: invoiceSettleTime,
			Amount:    invoiceOverpaidAmt,
			FiatValue: fiat,
			TxID:      invoiceHash,
			Reference: invoicePreimage,
			Note:      note,
			Type:      EntryTypeReceipt,
			OnChain:   false,
			Credit:    true,
		}

		if circular {
			expectedEntry.Type = EntryTypeCircularReceipt
		}

		return expectedEntry
	}
	tests := []struct {
		name     string
		circular bool
	}{
		{
			name:     "regular receive",
			circular: false,
		},

		{
			name:     "circular",
			circular: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			entry, err := invoiceEntry(
				invoice, test.circular, mockConvert,
			)
			if err != nil {
				t.Fatal(err)
			}

			expectedEntry := getEntry(test.circular)
			require.Equal(t, expectedEntry, entry)
		})
	}
}

// TestPaymentEntry tests creation of payment entries for circular rebalances
// and regular payments.
func TestPaymentEntry(t *testing.T) {
	// getEntries is a helper function which returns our expected entries
	// based on whether we are testing a payment to ourselves or not.
	getEntries := func(toSelf bool) []*HarmonyEntry {
		mockFiat, _ := mockConvert(int64(paymentMsat), paymentTime)
		paymentRef := paymentReference(
			uint64(paymentIndex), pmtHash,
		)

		paymentEntry := &HarmonyEntry{
			Timestamp: paymentTime,
			Amount:    lnwire.MilliSatoshi(paymentMsat),
			FiatValue: mockFiat,
			TxID:      paymentHash,
			Reference: paymentRef,
			Note:      paymentNote(pmtPreimage),
			Type:      EntryTypePayment,
			OnChain:   false,
			Credit:    false,
		}

		feeFiat, _ := mockConvert(int64(paymentFeeMsat), paymentTime)
		feeEntry := &HarmonyEntry{
			Timestamp: paymentTime,
			Amount:    lnwire.MilliSatoshi(paymentFeeMsat),
			FiatValue: feeFiat,
			TxID:      paymentHash,
			Reference: feeReference(paymentRef),
			Note:      paymentFeeNote(payment.Htlcs),
			Type:      EntryTypeFee,
			OnChain:   false,
			Credit:    false,
		}

		if toSelf {
			paymentEntry.Type = EntryTypeCircularPayment
			feeEntry.Type = EntryTypeCircularPaymentFee
		}

		return []*HarmonyEntry{paymentEntry, feeEntry}
	}

	tests := []struct {
		name   string
		toSelf bool
	}{
		{
			name:   "regular payment",
			toSelf: false,
		},
		{
			name:   "to self",
			toSelf: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			entries, err := paymentEntry(
				settledPmt, test.toSelf, mockConvert,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedEntries := getEntries(test.toSelf)

			require.Equal(t, expectedEntries, entries)
		})
	}
}

// TestForwardingEntry tests creation of a forwarding and forwarding fee entry.
func TestForwardingEntry(t *testing.T) {
	entries, err := forwardingEntry(fwdEntry, mockConvert)
	require.NoError(t, err)

	txid := forwardTxid(fwdEntry)
	note := forwardNote(fwdInMsat, fwdOutMsat)

	fwdFiat, _ := mockConvert(int64(0), forwardTs)
	fwdEntry := &HarmonyEntry{
		Timestamp: forwardTs,
		Amount:    0,
		FiatValue: fwdFiat,
		TxID:      txid,
		Reference: "",
		Note:      note,
		Type:      EntryTypeForward,
		OnChain:   false,
		Credit:    true,
	}

	feeFiat, _ := mockConvert(int64(fwdFeeMsat), forwardTs)

	feeEntry := &HarmonyEntry{
		Timestamp: forwardTs,
		Amount:    fwdFeeMsat,
		FiatValue: feeFiat,
		TxID:      txid,
		Reference: "",
		Note:      "",
		Type:      EntryTypeForwardFee,
		OnChain:   false,
		Credit:    true,
	}

	expectedEntries := []*HarmonyEntry{fwdEntry, feeEntry}
	require.Equal(t, expectedEntries, entries)
}
