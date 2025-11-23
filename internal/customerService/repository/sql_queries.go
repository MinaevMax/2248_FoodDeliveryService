package repository

const (
	createCustomer = `INSERT INTO customers VALUES(DEFAULT) RETURNING *`

	updateCustomer = `UPDATE customers SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING *`

	deleteCustomer = `DELETE FROM customers WHERE id = $1`

	getCustomerById = `SELECT * from customers WHERE id = $1`
)
