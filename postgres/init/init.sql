CREATE TABLE persons (
		employee_no INT PRIMARY KEY,
		dept_no INT,
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		email VARCHAR(100)
		);

INSERT INTO persons(employee_no, dept_no, first_name, last_name, email) VALUES
			(1, 10, 'Evan', 'MacMans', 'evanmacmans@example.com'),
			(2, 11, 'Malvina', 'FitzSimons', 'malvinafitzsimons@example.com'),
			(3, 12, 'Jimmie', 'Bruce', 'jimmiebruce@example.com')
			;
