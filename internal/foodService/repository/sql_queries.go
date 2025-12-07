package repository

const (
	// Order queries
	createOrderQuery = `INSERT INTO orders (user_id, updated_at)
		VALUES (:user_id, :updated_at) RETURNING id`

	createOrderQuery = `INSERT INTO orders (user_id) VALUES ($1) RETURNING id`
	
	getOrdersQuery = `SELECT id, status, updated_at FROM orders WHERE customer_id = $1 ORDER BY created_at DESC`

	getActiveOrdersQuery = `SELECT id, status, updated_at FROM orders WHERE user_id = $1 AND (status = UNDEFINED OR status = PACKING OR status = ARRIVING) ORDER BY created_at DESC`

	changeOrderStatusQuery = `UPDATE orders SET status = :status WHERE id = :order_id`

	// User queries
	registerUserQuery = `INSERT INTO users (login, password, created_at) VALUES ($1, $2, NOW()) RETURNING id`

	getUserByLoginQuery = `SELECT id, login, password, created_at FROM users WHERE login = $1`

	getUserByIDQuery = `SELECT id, login, password, email, phone, is_active, created_at, updated_at FROM users WHERE id = $1`

	createUserQuery = `INSERT INTO users (login, password, email, phone, is_active, created_at, updated_at) VALUES (:login, :password, :email, :phone, :is_active, NOW(), NOW()) RETURNING id, login, password, email, phone, is_active, created_at, updated_at`

	updateUserQuery = `UPDATE users SET login = :login, email = :email, phone = :phone, is_active = :is_active, updated_at = NOW() WHERE id = :id RETURNING id, login, password, email, phone, is_active, created_at, updated_at`

	deleteUserQuery = `DELETE FROM users WHERE id = $1`

	// Session queries
	createSessionQuery = `INSERT INTO sessions (session_id, user_id, created_at, expires_at) VALUES ($1, $2, NOW(), NOW() + INTERVAL '24 hours')`

	getSessionByUserIDQuery = `SELECT session_id, user_id, created_at, expires_at FROM sessions WHERE user_id = $1`

	updateSessionExpiryQuery = `UPDATE sessions SET expires_at = NOW() + INTERVAL '24 hours' WHERE session_id = $1`
)
