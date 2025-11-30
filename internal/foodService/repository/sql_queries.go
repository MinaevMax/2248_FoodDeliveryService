package repository

const (
	/* createCustomerQuery = `INSERT INTO customers (first_name, last_name, email, phone_number, password_hash, sex, birthday)
	VALUES (:first_name, :last_name, :email, :phone_number, :password_hash, :sex, :birthday)` */

	createOrderQuery = `INSERT INTO orders (customer_id, courier_id)
		VALUES (:customer_id, :courier_id) RETURNING id`
	
	getOrdersQuery = `SELECT id, status, updated_at FROM orders WHERE customer_id = $1 ORDER BY created_at DESC`

	ChangeOrderStatusQuery = `UPDATE orders SET status :status WHERE id = :order_id`

	createCustomerQuery = `INSERT INTO customers VALUES(DEFAULT) RETURNING *`

	updateCustomerQuery = `UPDATE customers SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING *`

	deleteCustomerQuery = `DELETE FROM customers WHERE id = $1`

	getCustomerByIdQuery = `SELECT * from customers WHERE id = $1`
)
