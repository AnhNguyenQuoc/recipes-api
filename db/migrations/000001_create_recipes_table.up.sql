CREATE TABLE IF NOT EXISTS recipes(
    id bigint PRIMARY KEY,
    name varchar(255),
    tags text[],
    ingredients text[],
    instructions text[],
    published_at timestamp
)