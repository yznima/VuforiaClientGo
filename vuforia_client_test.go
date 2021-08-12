package vuforia_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yznima/vuforia-client-go"
)

var (
	secretKey = os.Getenv("VUFORIA_SECRET_KEY")
	accessKey = os.Getenv("VUFORIA_ACCESS_KEY")
)

func TestTargetCRUD(t *testing.T) {
	t.Run("Target CRUD", func(t *testing.T) {
		client, err := vuforia.NewClient(vuforia.ClientConfig{
			SecretKey: secretKey,
			AccessKey: accessKey,
		})
		require.NoError(t, err)

		artWork, err := ioutil.ReadFile("./images/europeana-MvR30qxn-MM-unsplash.jpg")
		require.NoError(t, err)

		var targetId string
		t.Run("Post & Get", func(t *testing.T) {
			name := newTargetName()
			width := float64(2)
			active := true
			metadata := base64.RawStdEncoding.EncodeToString([]byte("License: Free to use under the Unsplash License"))
			resp, err := client.PostTarget(&vuforia.PostTargetRequest{
				Name:     name,
				Width:    width,
				Image:    base64.RawStdEncoding.EncodeToString(artWork),
				Active:   &active,
				Metadata: &metadata,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, resp.ResultCode, "TargetCreated")
			require.NotEmpty(t, resp.TargetId)
			targetId = resp.TargetId
			require.NotEmpty(t, resp.TransactionId)

			success := false
			for !success {
				getResp, err := client.GetTarget(&vuforia.GetTargetRequest{
					TargetId: resp.TargetId,
				})
				require.NoError(t, err)
				require.Equal(t, "Success", getResp.ResultCode)
				require.NotEmpty(t, getResp.TransactionId)
				require.Equal(t, active, getResp.TargetRecord.Active)
				require.Equal(t, name, getResp.TargetRecord.Name)
				require.Equal(t, width, getResp.TargetRecord.Width)
				require.True(t, getResp.Status == "processing" || getResp.Status == "success", getResp.Status)
				require.True(t, getResp.TargetRecord.TrackingRating == -1 || getResp.TargetRecord.TrackingRating == 5, getResp.TargetRecord.TrackingRating)
				success = getResp.Status == "success"
				if !success {
					time.Sleep(10 * time.Second)
				}
			}
		})

		t.Run("Update & Get", func(t *testing.T) {
			name2 := newTargetName()
			width2 := float64(3)
			image2 := base64.RawStdEncoding.EncodeToString(artWork)
			metadata2 := base64.RawStdEncoding.EncodeToString([]byte("New License: Free to use under the Unsplash License"))
			updateResp, err := client.UpdateTarget(&vuforia.UpdateTargetRequest{
				TargetId: targetId,
				Name:     &name2,
				Width:    &width2,
				Image:    &image2,
				Metadata: &metadata2,
			})
			require.NoError(t, err)
			require.Equal(t, "Success", updateResp.ResultCode)
			require.NotEmpty(t, updateResp.TransactionId)

			success := false
			for !success {
				getResp, err := client.GetTarget(&vuforia.GetTargetRequest{
					TargetId: targetId,
				})
				require.NoError(t, err)
				require.Equal(t, "Success", getResp.ResultCode)
				require.NotEmpty(t, getResp.TransactionId)
				require.Equal(t, true, getResp.TargetRecord.Active)
				require.Equal(t, name2, getResp.TargetRecord.Name)
				require.Equal(t, width2, getResp.TargetRecord.Width)
				require.True(t, getResp.Status == "processing" || getResp.Status == "success", getResp.Status)
				require.True(t, getResp.TargetRecord.TrackingRating == -1 || getResp.TargetRecord.TrackingRating == 5, getResp.TargetRecord.TrackingRating)
				success = getResp.Status == "success"
				if !success {
					time.Sleep(10 * time.Second)
				}
			}
		})

		t.Run("Summary", func(t *testing.T) {
			summaryResp, err := client.TargetSummary(&vuforia.TargetSummaryRequest{
				TargetId: targetId,
			})
			require.NoError(t, err)
			require.Equal(t, "Success", summaryResp.ResultCode)
			require.NotEmpty(t, summaryResp.TransactionId)
			require.Equal(t, true, summaryResp.Active)
			require.NotEmpty(t, summaryResp.TargetName)
			require.NotEmpty(t, summaryResp.DatabaseName)
			require.NotEmpty(t, summaryResp.UploadDate)
			require.Equal(t, summaryResp.TrackingRating, 5)
			require.Equal(t, summaryResp.TotalRecos, 0)
			require.Equal(t, summaryResp.CurrentMonthRecos, 0)
			require.Equal(t, summaryResp.PreviousMonthRecos, 0)
			require.Equal(t, "success", summaryResp.Status)
		})

		t.Run("Delete & Get", func(t *testing.T) {
			deleteResp, err := client.DeleteTarget(&vuforia.DeleteTargetRequest{
				TargetId: targetId,
			})
			require.NoError(t, err)
			require.Equal(t, "Success", deleteResp.ResultCode)
			require.NotEmpty(t, deleteResp.TransactionId)

			getResp, err := client.GetTarget(&vuforia.GetTargetRequest{
				TargetId: targetId,
			})
			require.NoError(t, err)
			require.Equal(t, "UnknownTarget", getResp.ResultCode)
		})
	})
}

// Have to create a new random name to avoid TargetNameExist error in case cleanup fails
func newTargetName() string {
	return "Target-" + fmt.Sprint(rand.Int())
}
