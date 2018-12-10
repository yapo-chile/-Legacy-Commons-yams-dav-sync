CREATE TABLE IF NOT EXISTS sync_error (
	sync_error_id        SERIAL PRIMARY KEY,
	image_path           VARCHAR(20) NOT NULL,
	error_counter    	 INT NOT NULL

);

ALTER TABLE sync_error ADD CONSTRAINT image_path_unique UNIQUE (image_path);


CREATE TABLE IF NOT EXISTS  last_sync (
    last_sync_id        SERIAL PRIMARY KEY,
    last_sync_date      TIMESTAMP
);

