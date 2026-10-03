-- ============================================
-- Be-Crustea Database Initialization
-- Migration: 000001_init.up.sql
-- ============================================

-- ============================================
-- ENUMS
-- ============================================

CREATE TYPE news_status AS ENUM (
    'draft',
    'published'
);

CREATE TYPE faq_category AS ENUM (
    'product',
    'service',
    'partnership'
);

-- ============================================
-- USERS
-- ============================================

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- ============================================
-- NEWS CATEGORIES
-- ============================================

CREATE TABLE news_categories (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name_idn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- ============================================
-- NEWS
-- ============================================

CREATE TABLE news (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    author_id BIGINT NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    thumbnail VARCHAR(255),
    title_idn VARCHAR(255) NOT NULL,
    title_en VARCHAR(255) NOT NULL,
    content_idn TEXT NOT NULL,
    content_en TEXT NOT NULL,
    status news_status NOT NULL DEFAULT 'draft',
    published_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_news_category
        FOREIGN KEY (category_id)
        REFERENCES news_categories(id),

    CONSTRAINT fk_news_author
        FOREIGN KEY (author_id)
        REFERENCES users(id)
);

CREATE INDEX idx_news_category_id
    ON news(category_id);

CREATE INDEX idx_news_author_id
    ON news(author_id);

-- ============================================
-- PRODUCT CATEGORIES
-- ============================================

CREATE TABLE product_categories (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name_idn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- ============================================
-- PRODUCTS
-- ============================================

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name_idn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    description_idn TEXT NOT NULL,
    description_en TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id)
        REFERENCES product_categories(id)
);

CREATE INDEX idx_products_category_id
    ON products(category_id);

-- ============================================
-- PRODUCT IMAGES
-- ============================================

CREATE TABLE product_images (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    file_url VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_product_images_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
);

CREATE INDEX idx_product_images_product_id
    ON product_images(product_id);

-- ============================================
-- PRODUCT EXCELLENCES
-- ============================================

CREATE TABLE product_excellences (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    title_idn VARCHAR(255) NOT NULL,
    title_en VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_product_excellences_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
);

CREATE INDEX idx_product_excellences_product_id
    ON product_excellences(product_id);

-- ============================================
-- PRODUCT SPECIFICATIONS
-- ============================================

CREATE TABLE product_specifications (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    name_idn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_product_specifications_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
);

CREATE INDEX idx_product_specifications_product_id
    ON product_specifications(product_id);

-- ============================================
-- FAQS
-- ============================================

CREATE TABLE faqs (
    id BIGSERIAL PRIMARY KEY,
    category faq_category NOT NULL,
    question_idn VARCHAR(255) NOT NULL,
    question_en VARCHAR(255) NOT NULL,
    answer_idn TEXT NOT NULL,
    answer_en TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_faqs_category
    ON faqs(category);

-- ============================================
-- PARTNERSHIP PROJECTS
-- ============================================

CREATE TABLE partnership_projects (
    id BIGSERIAL PRIMARY KEY,
    name_idn VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    description_idn TEXT NOT NULL,
    description_en TEXT NOT NULL,
    image VARCHAR(255),

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- ============================================
-- PARTNERSHIP PROJECT PARTNERS
-- ============================================

CREATE TABLE partnership_project_partners (
    id BIGSERIAL PRIMARY KEY,
    partnership_project_id BIGINT NOT NULL,
    logo VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_partnership_project_partners_project
        FOREIGN KEY (partnership_project_id)
        REFERENCES partnership_projects(id)
);

CREATE INDEX idx_partnership_project_partners_project_id
    ON partnership_project_partners(partnership_project_id);