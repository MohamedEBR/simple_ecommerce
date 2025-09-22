package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil{
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer	cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

// CART OPS 🛒

func CreateCart(ctx context.Context, db *sql.DB, userID string) (string, error) {

	var id string

	err := db.QueryRowContext(
		ctx,`
		INSERT INTO carts (user_id) VALUES ($1)
		RETURNING id
	`, userID).Scan(&id)
	
	return id, err
}

func AddItem(ctx context.Context, db *sql.DB, cartID, productID string, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be > 0")
	}

	_, err := db.ExecContext(
		ctx, `
			INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity,
		              updated_at = NOW()
	`, cartID, productID, qty)
	
	return err
}

type CartRow struct {
	ItemID      string  `json:"item_id"`
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	ProductName string  `json:"product_name"`
	PriceCents  int     `json:"price_cents"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func GetCart(ctx context.Context, db *sql.DB, cartID string) ([]CartRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ci.id, ci.product_id, ci.quantity, p.name, p.price_cents, ci.created_at, ci.updated_at
		FROM cart_items ci
		INNER JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at ASC
	`, cartID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []CartRow

	for rows.Next() {
		var r CartRow

		if err := rows.Scan(&r.ItemID, &r.ProductID, &r.Quantity, &r.ProductName, &r.PriceCents, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

func EmptyCart(ctx context.Context, db *sql.DB, cartID string) error {
	if _, err := db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, cartID); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
	UPDATE carts
	SET updated_at = NOW()
	WHERE id = $1
	`, cartID); err != nil {
		return err
	}

	return nil
}
func RemoveItem(ctx context.Context, db *sql.DB, cartID, productID string) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2
		`, cartID, productID)
	
	return err
}

func DecreaseItem(ctx context.Context, db *sql.DB, cartID, productID string, qty int) error {
	_, err := db.ExecContext(ctx, `
		UPDATE cart_items
		SET quantity = GREATEST(quantity - $3, 0), updated_at = NOW()
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID, qty)

	_, _ = db.ExecContext(ctx, `
		DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2 AND quantity = 0
	`, cartID, productID)

	return err
}