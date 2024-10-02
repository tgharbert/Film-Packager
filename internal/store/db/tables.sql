CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    password VARCHAR(255),
    role VARCHAR(50)
);

CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE memberships (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    organization_id INT REFERENCES organizations(id),
    access_tier VARCHAR(50)
);

-- CLI TO RUN THIS FROM THE ROOT DIR -- ADJUST PATH IF NECESSARY
-- NEEDED TO NAVIGATE INTO THE todos DB TO DO THIS...
-- \i internal/store/db/tables.sql