CREATE TABLE persons (
		employee_no INT PRIMARY KEY,
		dept_no INT,
		first_name CHAR(100),
		last_name CHAR(100),
		email CHAR(100)
		);

INSERT INTO persons(employee_no, dept_no, first_name, last_name, email) VALUES
			(1, 10, 'Evan', 'MacMans', 'evanmacmans@example.com'),
			(2, 11, 'Malvina', 'FitzSimons', 'malvinafitzsimons@examp.com'),
			(3, 12, 'Jimmie', 'Bruce', 'jimmiebruce@examp.com')
			;
