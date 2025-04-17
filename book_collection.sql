-- Create database
CREATE DATABASE IF NOT EXISTS book_collection CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE book_collection;

-- Drop tables if they exist
DROP TABLE IF EXISTS book;
DROP TABLE IF EXISTS author;
DROP TABLE IF EXISTS genre;
DROP TABLE IF EXISTS book_author;
DROP TABLE IF EXISTS book_genre;

-- Author table
CREATE TABLE author (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    bio TEXT
);

-- Genre table
CREATE TABLE genre (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- Book table
CREATE TABLE book (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    publication_year INT,
    status ENUM('TO_READ', 'READING', 'READ', 'ABANDONED') DEFAULT 'TO_READ',
    isbn VARCHAR(20) UNIQUE
);

-- Book genre join table
CREATE TABLE book_genre (
    book_id BIGINT,
    genre_id BIGINT,
    PRIMARY KEY (book_id, genre_id),
    FOREIGN KEY (book_id) REFERENCES book(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genre(id) ON DELETE CASCADE
);

-- Book author join table
CREATE TABLE book_author (
    book_id BIGINT,
    author_id BIGINT,
    PRIMARY KEY (book_id, author_id),
    FOREIGN KEY (book_id) REFERENCES book(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES author(id) ON DELETE CASCADE
);

CREATE INDEX idx_book_title ON book(title);
CREATE INDEX idx_book_read ON book(read_status);
CREATE INDEX idx_author_name ON author(name);

