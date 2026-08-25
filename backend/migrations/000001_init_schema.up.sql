CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT UNIQUE NOT NULL,
    shortcode TEXT UNIQUE,
    platform TEXT NOT NULL,
    raw_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name TEXT,
    instructions TEXT,
    difficulty TEXT,
    prep_time_minutes INT
);

CREATE TABLE ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE recipe_ingredients (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id UUID REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity TEXT,
    unit TEXT,
    PRIMARY KEY (recipe_id, ingredient_id)
);

CREATE TABLE recipe_tags (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    reasoning TEXT,
    PRIMARY KEY (recipe_id, tag)
);

CREATE TABLE recipe_embeddings (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    embedding vector(768) NOT NULL,
    PRIMARY KEY (recipe_id)
);

CREATE INDEX ON recipe_embeddings USING hnsw (embedding vector_cosine_ops);
