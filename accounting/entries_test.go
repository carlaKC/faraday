package accounting

import (
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/lightninglabs/faraday/fiat"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/routing/route"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var (
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
	onChainAmtMsat   = lnwire.MilliSatoshi(satsToMsat(onChainAmtSat))
	onChainFeeSat    = btcutil.Amount(1000)
	onChainFeeMsat   = lnwire.MilliSatoshi(satsToMsat(onChainFeeSat))
	onChainTimestamp = time.Unix(1588160816, 0)

	onChainReceive = lndclient.Transaction{
		TxHash:    onChainTxID,
		Amount:    onChainAmtSat,
		Timestamp: onChainTimestamp,
	}

	onChainSend = lndclient.Transaction{
		TxHash:    onChainTxID,
		Amount:    -onChainAmtSat - onChainFeeSat,
		Timestamp: onChainTimestamp,
		Fee:       onChainFeeSat,
	}

	sweep = lndclient.Transaction{
		TxHash:    onChainTxID,
		Amount:    onChainAmtSat,
		Timestamp: onChainTimestamp,
		Fee:       0,
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

	payInfo = paymentInfo{
		Payment:     payment,
		destination: &otherPubkey,
		settleTime:  paymentTime,
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

	mockPriceTimestamp = time.Unix(1594306589, 0)

	mockBTCPrice = &fiat.USDPrice{
		Timestamp: mockPriceTimestamp,
		Price:     decimal.NewFromInt(100000),
	}
)

// mockPrice is a mocked price function which returns mockPrice * amount.
func mockPrice(_ time.Time) (*fiat.USDPrice, error) {
	return mockBTCPrice, nil
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

		note := channelOpenNote(
			initiator, remotePubkey, channelCapacitySats,
		)

		amtMsat := lnwire.MilliSatoshi(amt)
		return &HarmonyEntry{
			Timestamp: transactionTimestamp,
			Amount:    amtMsat,
			FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, amtMsat),
			TxID:      openChannelTx,
			Reference: fmt.Sprintf("%v", channelID),
			Note:      note,
			Type:      entryType,
			OnChain:   true,
			Credit:    credit,
			BTCPrice:  mockBTCPrice,
		}

	}

	feeAmt := satsToMsat(channelFeesSats)
	msatAmt := lnwire.MilliSatoshi(feeAmt)

	// Fee entry is the expected fee entry for locally initiated channels.
	feeEntry := &HarmonyEntry{
		Timestamp: transactionTimestamp,
		Amount:    msatAmt,
		FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, msatAmt),
		TxID:      openChannelTx,
		Reference: feeReference(openChannelTx),
		Note:      channelOpenFeeNote(channelID),
		Type:      EntryTypeChannelOpenFee,
		OnChain:   true,
		Credit:    false,
		BTCPrice:  mockBTCPrice,
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
				channel, tx, mockPrice,
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
		amtMsat := lnwire.MilliSatoshi(closeAmt)

		return &HarmonyEntry{
			Timestamp: closeTimestamp,
			Amount:    amtMsat,
			FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, amtMsat),
			TxID:      closeTx,
			Reference: closeTx,
			Note:      note,
			Type:      EntryTypeChannelClose,
			OnChain:   true,
			Credit:    true,
			BTCPrice:  mockBTCPrice,
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
				closeChan, closeTx, mockPrice,
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

// TestSweepEntry tests creation of entries for a sweep.
func TestSweepEntry(t *testing.T) {
	sweepWithOutFee := sweep
	sweepWithFee := sweep
	sweepWithFee.Fee = onChainFeeSat

	expectedSweep := &HarmonyEntry{
		Timestamp: onChainTimestamp,
		Amount:    onChainAmtMsat,
		FiatValue: fiat.MsatToUSD(
			mockBTCPrice.Price,
			onChainAmtMsat,
		),
		TxID:      onChainTxID,
		Reference: onChainTxID,
		Type:      EntryTypeSweep,
		OnChain:   true,
		Credit:    true,
		BTCPrice:  mockBTCPrice,
	}

	tests := []struct {
		name     string
		tx       lndclient.Transaction
		expected []*HarmonyEntry
	}{
		{
			name: "sweep with fee",
			tx:   sweepWithFee,
			expected: []*HarmonyEntry{
				expectedSweep,
				{
					Timestamp: onChainTimestamp,
					Amount:    onChainFeeMsat,
					FiatValue: fiat.MsatToUSD(
						mockBTCPrice.Price,
						onChainFeeMsat,
					),
					TxID:      onChainTxID,
					Reference: feeReference(onChainTxID),
					Note:      "",
					Type:      EntryTypeSweepFee,
					OnChain:   true,
					Credit:    false,
					BTCPrice:  mockBTCPrice,
				},
			},
		},
		{
			name: "sweep without fee",
			tx:   sweepWithOutFee,
			expected: []*HarmonyEntry{
				expectedSweep,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			entries, err := sweepEntry(test.tx, mockPrice)
			require.NoError(t, err)
			require.Equal(t, test.expected, entries)
		})
	}
}

// TestOnChainEntry tests creation of entries for receipts and payments, and the
// generation of a fee entry where applicable.
func TestOnChainEntry(t *testing.T) {
	onChainReceiveWithFee := onChainReceive
	onChainReceiveWithFee.Fee = onChainFeeSat

	tests := []struct {
		name string

		tx lndclient.Transaction

		expectedEntries []*HarmonyEntry

		err error
	}{
		{
			name: "receive without fee",
			tx:   onChainReceive,
			expectedEntries: []*HarmonyEntry{
				{
					Timestamp: onChainTimestamp,
					Amount:    onChainAmtMsat,
					FiatValue: fiat.MsatToUSD(
						mockBTCPrice.Price,
						onChainAmtMsat,
					),
					TxID:      onChainTxID,
					Reference: onChainTxID,
					Type:      EntryTypeReceipt,
					OnChain:   true,
					Credit:    true,
					BTCPrice:  mockBTCPrice,
				},
			},
			err: nil,
		},
		{
			name:            "receive with fee",
			tx:              onChainReceiveWithFee,
			expectedEntries: nil,
			err:             ErrReceiveWithFees,
		},
		{
			name: "payment with fee",
			tx:   onChainSend,
			expectedEntries: []*HarmonyEntry{
				{
					Timestamp: onChainTimestamp,
					Amount:    onChainAmtMsat,
					FiatValue: fiat.MsatToUSD(
						mockBTCPrice.Price,
						onChainAmtMsat,
					),
					TxID:      onChainTxID,
					Reference: onChainTxID,
					Type:      EntryTypePayment,
					OnChain:   true,
					Credit:    false,
					BTCPrice:  mockBTCPrice,
				},
				{
					Timestamp: onChainTimestamp,
					Amount:    onChainFeeMsat,
					FiatValue: fiat.MsatToUSD(
						mockBTCPrice.Price,
						onChainFeeMsat,
					),
					TxID:      onChainTxID,
					Reference: feeReference(onChainTxID),
					Note:      "",
					Type:      EntryTypeFee,
					OnChain:   true,
					Credit:    false,
					BTCPrice:  mockBTCPrice,
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			entries, err := onChainEntries(test.tx, mockPrice)
			require.Equal(t, test.err, err)

			require.Equal(t, test.expectedEntries, entries)
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

		expectedEntry := &HarmonyEntry{
			Timestamp: invoiceSettleTime,
			Amount:    invoiceOverpaidAmt,
			FiatValue: fiat.MsatToUSD(
				mockBTCPrice.Price, invoiceOverpaidAmt,
			),
			TxID:      invoiceHash,
			Reference: invoicePreimage,
			Note:      note,
			Type:      EntryTypeReceipt,
			OnChain:   false,
			Credit:    true,
			BTCPrice:  mockBTCPrice,
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
				invoice, test.circular, mockPrice,
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
		paymentRef := paymentReference(
			uint64(paymentIndex), pmtPreimage,
		)

		amtMsat := lnwire.MilliSatoshi(paymentMsat)

		paymentEntry := &HarmonyEntry{
			Timestamp: paymentTime,
			Amount:    amtMsat,
			FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, amtMsat),
			TxID:      paymentHash,
			Reference: paymentRef,
			Note:      paymentNote(&otherPubkey),
			Type:      EntryTypePayment,
			OnChain:   false,
			Credit:    false,
			BTCPrice:  mockBTCPrice,
		}

		feeMsat := lnwire.MilliSatoshi(paymentFeeMsat)

		feeEntry := &HarmonyEntry{
			Timestamp: paymentTime,
			Amount:    feeMsat,
			FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, feeMsat),
			TxID:      paymentHash,
			Reference: feeReference(paymentRef),
			Note:      paymentNote(&otherPubkey),
			Type:      EntryTypeFee,
			OnChain:   false,
			Credit:    false,
			BTCPrice:  mockBTCPrice,
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
				payInfo, test.toSelf, mockPrice,
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
	entries, err := forwardingEntry(fwdEntry, mockPrice)
	require.NoError(t, err)

	txid := forwardTxid(fwdEntry)
	note := forwardNote(fwdInMsat, fwdOutMsat)

	fwdEntry := &HarmonyEntry{
		Timestamp: forwardTs,
		Amount:    0,
		FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, 0),
		TxID:      txid,
		Reference: "",
		Note:      note,
		Type:      EntryTypeForward,
		OnChain:   false,
		Credit:    true,
		BTCPrice:  mockBTCPrice,
	}

	feeEntry := &HarmonyEntry{
		Timestamp: forwardTs,
		Amount:    fwdFeeMsat,
		FiatValue: fiat.MsatToUSD(mockBTCPrice.Price, fwdFeeMsat),
		TxID:      txid,
		Reference: "",
		Note:      "",
		Type:      EntryTypeForwardFee,
		OnChain:   false,
		Credit:    true,
		BTCPrice:  mockBTCPrice,
	}

	expectedEntries := []*HarmonyEntry{fwdEntry, feeEntry}
	require.Equal(t, expectedEntries, entries)
}
