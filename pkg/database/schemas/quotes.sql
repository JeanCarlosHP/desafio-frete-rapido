CREATE TABLE quotes (
  id SERIAL PRIMARY KEY,
  carrier_name VARCHAR(255) NOT NULL,
  service VARCHAR(255) NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  deadline INT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);