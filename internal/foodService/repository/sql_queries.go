package repository

const createCustomerQuery = `INSERT INTO customers (first_name, last_name, email, phone_number, password_hash, sex, birthday)
		VALUES (:first_name, :last_name, :email, :phone_number, :password_hash, :sex, :birthday)`

const newOrderQuery = `INSERT INTO orders (customer_id, courier_id, amount)
		VALUES (:customer_id, :courier_id, :amount)`
