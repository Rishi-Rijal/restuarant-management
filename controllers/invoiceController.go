package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rishi-rijal/restuarant-management/database"
	"github.com/rishi-rijal/restuarant-management/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type InvoiceViewFormat struct {
	InvoiceID      string
	PaymentMethod  string
	OrderID        string
	PaymentStatus  *string
	PaymentDue     interface{}
	TableNumber    interface{}
	PaymentDueDate time.Time
	OrderDetails   interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := invoiceCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the invoices"})
			return
		}

		var allInvoices []bson.M

		if err := result.All(ctx, allInvoices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong while fetching invoices"})
			return
		}

		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		invoiceID := c.Param("invoice_id")
		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx, bson.M{"invoiceID": invoiceID}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching the invoice"})
			return
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.OrderID)
		invoiceView.OrderID = invoice.OrderID
		invoiceView.PaymentDueDate = invoice.PaymentDueDate
		invoiceView.PaymentMethod = "null"
		if invoice.PaymentMethod != nil {
			invoiceView.PaymentMethod = *invoice.PaymentMethod
		}

		invoiceView.InvoiceID = invoice.InvoiceID
		invoiceView.PaymentStatus = *&invoice.PaymentStatus
		invoiceView.PaymentDue = allOrderItems[0]["PaymentDue"]
		invoiceView.TableNumber = allOrderItems[0]["TableNumber"]
		invoiceView.OrderDetails = allOrderItems[0]["orderItems"]

		c.JSON(http.StatusOK, invoice)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"orderID": invoice.OrderID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order not found"})
			return
		}

		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}

		invoice.PaymentDueDate = time.Now().AddDate(0, 6, 0) // 6 months form now

		invoice.CreatedAt = time.Now()
		invoice.UpdatedAt = time.Now()
		invoice.ID = bson.NewObjectID()

		invoice.InvoiceID = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(ctx, invoice)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice item not created"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var invoice models.Invoice

		invoiceID := c.Param("invoice_id")

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"invoiceID": invoiceID}

		var updateObj bson.D

		if invoice.PaymentMethod != nil {
			updateObj = append(updateObj, bson.E{"PaymentMethod", invoice.PaymentMethod})
		}

		if invoice.PaymentStatus != nil {
			updateObj = append(updateObj, bson.E{"PaymentStatus", invoice.PaymentStatus})
		}

		invoice.UpdatedAt = time.Now()
		updateObj = append(updateObj, bson.E{"UpdatedAt", invoice.UpdatedAt})

		opt := options.UpdateOne().SetUpsert(true)

		status := "PENDING"

		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}

		result, err := invoiceCollection.UpdateOne(
			ctx, filter, bson.D{
				{"$set", updateObj},
			},

			opt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invoice update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
