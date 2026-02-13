// +build integration

package pix

import (
	"context"
	"testing"
	"time"

	"github.com/pericles-luz/go-bb-pix/bbpix"
	"github.com/pericles-luz/go-bb-pix/internal/testutil"
)

func init() {
	// Load .env file for integration tests
	_ = testutil.LoadEnv()
}

// TestE2E_CompleteQRCodeFlow tests the complete QR Code lifecycle
func TestE2E_CompleteQRCodeFlow(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Check if credentials are available
	if !testutil.HasCredentials() {
		t.Skip("Integration test credentials not configured. Create .env file from .env.example")
	}

	// Load credentials from environment
	envConfig := testutil.GetBBConfig()
	config := bbpix.Config{
		Environment:     envConfig["environment"],
		ClientID:        envConfig["client_id"],
		ClientSecret:    envConfig["client_secret"],
		DeveloperAppKey: envConfig["dev_app_key"],
	}

	// Create client
	client, err := bbpix.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pixClient := client.PIX()
	ctx := context.Background()

	// Generate unique txid
	txid := generateUniqueTxID()

	// Step 1: Create QR Code
	t.Log("Creating QR Code...")
	createReq := CreateQRCodeRequest{
		TxID:       txid,
		Value:      37.00,
		Expiration: 3600,
		Debtor: &Debtor{
			CPF:  "12345678909",
			Name: "Francisco da Silva",
		},
		PayerSolicitation: "Teste de integração",
	}

	qrCode, err := pixClient.CreateQRCode(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create QR Code: %v", err)
	}

	if qrCode.TxID != txid {
		t.Errorf("TxID = %s, want %s", qrCode.TxID, txid)
	}

	if qrCode.Status != "ATIVA" {
		t.Errorf("Status = %s, want ATIVA", qrCode.Status)
	}

	if qrCode.QRCode == "" {
		t.Error("QRCode should not be empty")
	}

	t.Logf("QR Code created successfully: %s", qrCode.TxID)

	// Step 2: Get QR Code
	t.Log("Retrieving QR Code...")
	retrieved, err := pixClient.GetQRCode(ctx, txid)
	if err != nil {
		t.Fatalf("Failed to get QR Code: %v", err)
	}

	if retrieved.TxID != qrCode.TxID {
		t.Errorf("Retrieved TxID = %s, want %s", retrieved.TxID, qrCode.TxID)
	}

	if retrieved.Revision != qrCode.Revision {
		t.Errorf("Retrieved Revision = %d, want %d", retrieved.Revision, qrCode.Revision)
	}

	t.Logf("QR Code retrieved successfully: %s", retrieved.TxID)

	// Step 3: Update QR Code
	t.Log("Updating QR Code...")
	updateReq := UpdateQRCodeRequest{
		Value:      50.00,
		Expiration: 7200,
	}

	updated, err := pixClient.UpdateQRCode(ctx, txid, updateReq)
	if err != nil {
		t.Fatalf("Failed to update QR Code: %v", err)
	}

	if updated.Value.Original != "50.00" {
		t.Errorf("Updated value = %s, want 50.00", updated.Value.Original)
	}

	if updated.Revision <= qrCode.Revision {
		t.Errorf("Revision should increase after update: %d -> %d", qrCode.Revision, updated.Revision)
	}

	t.Logf("QR Code updated successfully, new revision: %d", updated.Revision)

	// Step 4: List QR Codes
	t.Log("Listing QR Codes...")
	listParams := ListQRCodesParams{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
		CPF:       "12345678909",
	}

	list, err := pixClient.ListQRCodes(ctx, listParams)
	if err != nil {
		t.Fatalf("Failed to list QR Codes: %v", err)
	}

	found := false
	for _, qr := range list.QRCodes {
		if qr.TxID == txid {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created QR Code not found in list")
	}

	t.Logf("Listed %d QR Codes", len(list.QRCodes))

	// Step 5: Delete QR Code (cleanup)
	t.Log("Deleting QR Code...")
	err = pixClient.DeleteQRCode(ctx, txid)
	if err != nil {
		t.Errorf("Failed to delete QR Code: %v", err)
	}

	t.Log("QR Code deleted successfully")

	// Step 6: Verify deletion
	t.Log("Verifying deletion...")
	_, err = pixClient.GetQRCode(ctx, txid)
	if err == nil {
		t.Error("Expected error when getting deleted QR Code")
	}

	t.Log("Complete flow finished successfully")
}

// TestE2E_PaymentAndRefundFlow tests payment reception and refund
func TestE2E_PaymentAndRefundFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !testutil.HasCredentials() {
		t.Skip("Integration test credentials not configured. Create .env file from .env.example")
	}

	envConfig := testutil.GetBBConfig()
	config := bbpix.Config{
		Environment:     envConfig["environment"],
		ClientID:        envConfig["client_id"],
		ClientSecret:    envConfig["client_secret"],
		DeveloperAppKey: envConfig["dev_app_key"],
	}

	client, err := bbpix.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pixClient := client.PIX()
	ctx := context.Background()

	// Step 1: List recent payments
	t.Log("Listing recent payments...")
	params := ListPaymentsParams{
		StartDate: time.Now().Add(-5 * 24 * time.Hour),
		EndDate:   time.Now(),
		PageSize:  10,
	}

	payments, err := pixClient.ListPayments(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list payments: %v", err)
	}

	t.Logf("Found %d payments", len(payments.Payments))

	if len(payments.Payments) == 0 {
		t.Skip("No payments available for refund test")
		return
	}

	// Step 2: Get payment details
	payment := payments.Payments[0]
	t.Logf("Getting payment details: %s", payment.EndToEndID)

	paymentDetail, err := pixClient.GetPayment(ctx, payment.EndToEndID)
	if err != nil {
		t.Fatalf("Failed to get payment: %v", err)
	}

	if paymentDetail.EndToEndID != payment.EndToEndID {
		t.Errorf("EndToEndID mismatch: %s != %s", paymentDetail.EndToEndID, payment.EndToEndID)
	}

	// Step 3: Create refund (if payment value allows)
	// Note: In sandbox, this might not create actual refunds
	// This is mainly to test the API integration
	t.Log("Creating refund...")
	refundID := generateUniqueRefundID()
	refundReq := CreateRefundRequest{
		Value:  5.00,
		Reason: "Teste de integração",
	}

	refund, err := pixClient.CreateRefund(ctx, payment.EndToEndID, refundID, refundReq)
	if err != nil {
		// In sandbox, refunds might fail - log but don't fail test
		t.Logf("Refund creation failed (expected in sandbox): %v", err)
		return
	}

	if refund.ID != refundID {
		t.Errorf("Refund ID = %s, want %s", refund.ID, refundID)
	}

	t.Logf("Refund created: %s, status: %s", refund.ID, refund.Status)

	// Step 4: Get refund status
	t.Log("Checking refund status...")
	refundStatus, err := pixClient.GetRefund(ctx, payment.EndToEndID, refundID)
	if err != nil {
		t.Errorf("Failed to get refund status: %v", err)
		return
	}

	if refundStatus.ID != refundID {
		t.Errorf("Refund status ID = %s, want %s", refundStatus.ID, refundID)
	}

	t.Logf("Refund status: %s", refundStatus.Status)
}

// TestE2E_PaginationHandling tests pagination in list operations
func TestE2E_PaginationHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !testutil.HasCredentials() {
		t.Skip("Integration test credentials not configured. Create .env file from .env.example")
	}

	envConfig := testutil.GetBBConfig()
	config := bbpix.Config{
		Environment:     envConfig["environment"],
		ClientID:        envConfig["client_id"],
		ClientSecret:    envConfig["client_secret"],
		DeveloperAppKey: envConfig["dev_app_key"],
	}

	client, err := bbpix.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pixClient := client.PIX()
	ctx := context.Background()

	// Test pagination with small page size
	t.Log("Testing pagination...")
	params := ListQRCodesParams{
		StartDate: time.Now().Add(-30 * 24 * time.Hour),
		EndDate:   time.Now(),
		PageSize:  5, // Small page to test pagination
		Page:      0,
	}

	firstPage, err := pixClient.ListQRCodes(ctx, params)
	if err != nil {
		t.Fatalf("Failed to get first page: %v", err)
	}

	t.Logf("First page: %d items, total: %d, pages: %d",
		len(firstPage.QRCodes),
		firstPage.Parameters.Pagination.TotalItems,
		firstPage.Parameters.Pagination.TotalPages)

	if firstPage.Parameters.Pagination.TotalPages > 1 {
		// Fetch second page
		params.Page = 1
		secondPage, err := pixClient.ListQRCodes(ctx, params)
		if err != nil {
			t.Fatalf("Failed to get second page: %v", err)
		}

		t.Logf("Second page: %d items", len(secondPage.QRCodes))

		// Verify pages are different
		if len(firstPage.QRCodes) > 0 && len(secondPage.QRCodes) > 0 {
			if firstPage.QRCodes[0].TxID == secondPage.QRCodes[0].TxID {
				t.Error("Pages contain same items")
			}
		}
	}
}

// Helper functions

func generateUniqueTxID() string {
	// Generate a unique txid (26-35 characters)
	timestamp := time.Now().UnixNano()
	return "test" + string(rune(timestamp%1000000000000000000)) + "1234567890123456"
}

func generateUniqueRefundID() string {
	timestamp := time.Now().UnixNano()
	return "dev" + string(rune(timestamp%100000000))
}
