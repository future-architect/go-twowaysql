CREATE TABLE persons (
		employee_no INT PRIMARY KEY,
		dept_no INT,
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		email VARCHAR(100),
		null_string VARCHAR(100),
		null_int INT,
		created_at timestamp with time zone NOT NULL,
		updated_at timestamp with time zone
		);

INSERT INTO persons(employee_no, dept_no, first_name, last_name, email, created_at) VALUES
			(1, 10, 'Evan', 'MacMans', 'evanmacmans@example.com', CURRENT_TIMESTAMP),
			(2, 11, 'Malvina', 'FitzSimons', 'malvinafitzsimons@example.com', CURRENT_TIMESTAMP),
			(3, 12, 'Jimmie', 'Bruce', 'jimmiebruce@example.com', CURRENT_TIMESTAMP)
			;
