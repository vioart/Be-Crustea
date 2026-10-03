-- ============================================
-- Be-Crustea Database Rollback
-- Migration: 000001_init.down.sql
-- ============================================

-- ============================================
-- DROP TABLES
-- ============================================

DROP TABLE IF EXISTS partnership_project_partners;
DROP TABLE IF EXISTS partnership_projects;

DROP TABLE IF EXISTS faqs;

DROP TABLE IF EXISTS product_specifications;
DROP TABLE IF EXISTS product_excellences;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS product_categories;

DROP TABLE IF EXISTS news;
DROP TABLE IF EXISTS news_categories;

DROP TABLE IF EXISTS users;

-- ============================================
-- DROP ENUMS
-- ============================================

DROP TYPE IF EXISTS faq_category;
DROP TYPE IF EXISTS news_status;