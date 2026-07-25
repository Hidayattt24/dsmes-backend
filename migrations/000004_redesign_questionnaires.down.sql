-- Migration: 000004_redesign_questionnaires.down.sql

DROP TABLE IF EXISTS question_options;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS question_categories;
DROP TABLE IF EXISTS questionnaires;
DROP TYPE IF EXISTS questionnaire_type_enum;
