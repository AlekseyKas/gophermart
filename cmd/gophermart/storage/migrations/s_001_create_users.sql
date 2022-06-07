
CREATE TABLE users (
  user_id INT NOT NULL GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  login VARCHAR (50)  UNIQUE NOT NULL,
  password VARCHAR (100) NOT NULL,
  cookie jsonb
);

CREATE TABLE orders (
  order_id INT NOT NULL GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id INT,
  UNIQUE (user_id, number),
  number VARCHAR ( 50 ) UNIQUE NOT NULL,
  status VARCHAR (50) NOT NULL DEFAULT 'NEW',
  accrual DOUBLE PRECISION DEFAULT 0,
  UploadedAt TIMESTAMPTZ,
  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(user_id)
);
CREATE TABLE balance (  
  balance_id INT NOT NULL GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id INT,
  balance DOUBLE PRECISION,
  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(user_id)
);

CREATE TABLE withdraws (  
  withdraw_id INT NOT NULL GENERATED ALWAYS AS IDENTITY,
  user_id INT,
  order VARCHAR ( 50 ) UNIQUE NOT NULL,
  withdraws DOUBLE PRECISION,
  processed_at TIMESTAMPTZ,
  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(user_id)
)
