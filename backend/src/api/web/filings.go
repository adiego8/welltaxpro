package webapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"welltaxpro/src/internal/notification"

	"github.com/google/logger"
	"github.com/gorilla/mux"
)

// markFilingCompleted marks a filing as completed (admin only)
func (api *API) markFilingCompleted(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenantId"]
	filingID := vars["filingId"]

	logger.Infof("Mark filing %s as completed for tenant %s", filingID, tenantID)

	// Get tenant database connection
	tenantDB, tc, err := api.store.GetTenantDB(tenantID)
	if err != nil {
		logger.Errorf("Failed to get tenant database: %v", err)
		http.Error(w, "Failed to connect to tenant database", http.StatusInternalServerError)
		return
	}

	// Update filing_status to mark as completed
	updateQuery := `
		UPDATE ` + tc.SchemaPrefix + `.filing_status
		SET is_completed = true, status = 'COMPLETED'
		WHERE filing_id = $1
	`

	result, err := tenantDB.Exec(updateQuery, filingID)
	if err != nil {
		logger.Errorf("Failed to update filing status: %v", err)
		http.Error(w, "Failed to update filing status", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf("Failed to get rows affected: %v", err)
		http.Error(w, "Failed to verify update", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Filing status not found", http.StatusNotFound)
		return
	}

	logger.Infof("Successfully marked filing %s as completed", filingID)

	// Get filing and client information for email notification
	var clientEmail, clientFirstName, clientLastName string
	var taxYear int
	var filingType string

	filingQuery := `
		SELECT
			u.email,
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			f.year,
			'Tax Return'
		FROM ` + tc.SchemaPrefix + `.filing f
		JOIN ` + tc.SchemaPrefix + `.user u ON f.user_id = u.id
		WHERE f.id = $1
	`

	err = tenantDB.QueryRow(filingQuery, filingID).Scan(
		&clientEmail,
		&clientFirstName,
		&clientLastName,
		&taxYear,
		&filingType,
	)

	if err != nil {
		logger.Warningf("Failed to get client info for email notification: %v", err)
		// Don't fail the request, just skip the email
	} else {
		// Send email notification
		clientName := clientFirstName
		if clientLastName != "" {
			clientName = fmt.Sprintf("%s %s", clientFirstName, clientLastName)
		}
		if clientName == "" {
			clientName = "Valued Client"
		}

		// Generate email content
		subject, htmlBody, textBody := notification.GenerateFilingCompletedEmail(notification.FilingCompletedEmail{
			ClientName: clientName,
			TaxYear:    taxYear,
			FilingType: filingType,
			TenantName: tc.TenantName,
			LoginURL:   fmt.Sprintf("https://app.welltaxpro.com/%s/clients", tenantID),
		})

		// Send email
		err = api.emailService.SendEmail(clientEmail, clientName, subject, htmlBody, textBody)
		if err != nil {
			logger.Errorf("Failed to send filing completed email to %s: %v", clientEmail, err)
			// Don't fail the request, email is not critical
		} else {
			logger.Infof("Filing completed email sent to %s", clientEmail)
		}
	}

	// Return success response
	response := map[string]interface{}{
		"status":      "COMPLETED",
		"isCompleted": true,
		"message":     "Filing marked as completed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("Failed to encode response: %v", err)
	}
}
