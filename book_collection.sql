-- Enable InnoDB engine for transactions and referential integrity
-- Already the default in modern MySQL versions
SET default_storage_engine = 'InnoDB';

-- Create database
CREATE DATABASE IF NOT EXISTS book_collection CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE book_collection;

-- Drop views and then drop tables in reverse dependency order
DROP VIEW IF EXISTS vw_books_detailed;
DROP VIEW IF EXISTS vw_user_books_detailed;
DROP TABLE IF EXISTS audit_log, user_book, book_request, user_role, role_permission, permission,
    role, book_author, book_genre, book, author, genre, user;

-- Author table
CREATE TABLE `author` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(45) NOT NULL,
    last_name VARCHAR(45) NOT NULL,
    description TEXT
);

-- Genre table
CREATE TABLE `genre` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Book table
CREATE TABLE `book` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    publication_year INT,
    isbn VARCHAR(13) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted BOOLEAN DEFAULT FALSE
);

-- Book author join table
CREATE TABLE `book_author` (
    book_id BIGINT,
    author_id BIGINT,
    PRIMARY KEY (book_id, author_id),
    FOREIGN KEY (book_id) REFERENCES book(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES author(id) ON DELETE CASCADE
);

-- Book genre join table
CREATE TABLE `book_genre` (
    book_id BIGINT,
    genre_id BIGINT,
    PRIMARY KEY (book_id, genre_id),
    FOREIGN KEY (book_id) REFERENCES book(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genre(id) ON DELETE CASCADE
);

-- User table
CREATE TABLE `user` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Role table
CREATE TABLE `role` (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

-- Permission table
CREATE TABLE `permission` (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

-- User role join table
CREATE TABLE `user_role` (
    user_id BIGINT,
    role_id INT,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE
);

-- Role permission join table
CREATE TABLE `role_permission` (
    role_id INT,
    permission_id INT,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permission(id) ON DELETE CASCADE
);

-- User book join table
CREATE TABLE user_book (
    user_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    status ENUM('TO_READ', 'READING', 'READ', 'ABANDONED') DEFAULT 'TO_READ',
    notes TEXT,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, book_id),
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES book(id) ON DELETE CASCADE
);

-- Book request table
CREATE TABLE book_request (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    author_name VARCHAR(255),
    isbn VARCHAR(13),
    status ENUM('PENDING', 'APPROVED', 'REJECTED') DEFAULT 'PENDING',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP NULL,
    reviewer_id BIGINT,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES user(id) ON DELETE SET NULL
);

-- Audit log table
CREATE TABLE `audit_log` (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    details TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE SET NULL
);

-- Triggers
DELIMITER //
CREATE TRIGGER after_book_insert AFTER INSERT ON book
    FOR EACH ROW BEGIN
        INSERT INTO audit_log (user_id, action, details)
        VALUES (NULL, 'BOOK_CREATE', CONCAT('Book "', NEW.title, '" added with ID ', NEW.id));
END;
//

CREATE TRIGGER after_book_update AFTER UPDATE ON book
    FOR EACH ROW BEGIN
        INSERT INTO audit_log (user_id, action, details)
        VALUES (NULL, 'BOOK_UPDATE', CONCAT('Book "', OLD.title, '" (ID ', OLD.id, ') was updated.'));
END;
//

CREATE TRIGGER after_book_delete AFTER DELETE ON book
    FOR EACH ROW BEGIN
        INSERT INTO audit_log (user_id, action, details)
        VALUES (NULL, 'BOOK_DELETE', CONCAT('Book "', OLD.title, '" (ID ', OLD.id, ') was deleted.'));
END;
//

CREATE TRIGGER after_user_insert AFTER INSERT ON user
    FOR EACH ROW BEGIN
        INSERT INTO audit_log (user_id, action, details)
        VALUES (NEW.id, 'USER_CREATE', CONCAT('User "', NEW.username, '" registered.'));
END;
//

CREATE TRIGGER after_user_role_insert AFTER INSERT ON user_role
    FOR EACH ROW BEGIN
        INSERT INTO audit_log(user_id, action, details)
        VALUES (NEW.user_id, 'ROLE_ASSIGNED',
                CONCAT('Role ID ', NEW.role_id, ' assigned to user ', NEW.user_id));
END;
//

CREATE TRIGGER after_book_request_insert
    AFTER INSERT ON book_request
    FOR EACH ROW
BEGIN
    INSERT INTO audit_log (user_id, action, details)
    VALUES (NEW.user_id, 'BOOK_REQUEST',
            CONCAT('User requested book "', NEW.title, '".'));
END;
//

CREATE TRIGGER after_book_request_update
    AFTER UPDATE ON book_request
    FOR EACH ROW
BEGIN
    IF OLD.status <> NEW.status THEN
        INSERT INTO audit_log (user_id, action, details)
        VALUES (NEW.reviewer_id, 'BOOK_REQUEST_REVIEW',
                CONCAT('Request for "', NEW.title, '" was ', NEW.status, '.'));
    END IF;
END;
//

DELIMITER ;

-- Insert roles
INSERT INTO role (id, name, description) VALUES
    (1, 'ADMIN', 'Administrator with full access'),
    (2, 'MODERATOR', 'Can review and approve content'),
    (3, 'USER', 'Regular user'),
    (4, 'GUEST', 'Read-only access');

-- Insert permissions
INSERT INTO permission (id, name) VALUES
    (1, 'BOOK_REQUEST'), (2, 'BOOK_CREATE'), (3, 'BOOK_UPDATE'), (4, 'BOOK_DELETE'),
    (5, 'BOOK_VIEW'), (6, 'USER_MANAGE'), (7, 'AUDIT_VIEW');

-- Map role to permission
INSERT INTO role_permission (role_id, permission_id) VALUES
    -- ADMIN -> ALL PERMISSIONS
    (1, 1), (1, 2), (1, 3), (1, 4),
    (1, 5), (1, 6), (1, 7),
    -- MODERATOR -> BOOK_VIEW, AUDIT_VIEW
    (2, 5), (2, 7),
    -- USER -> BOOK_REQUEST, BOOK_VIEW
    (3, 1), (3, 5),
    -- GUEST -> BOOK_VIEW
    (4, 5);

-- Useful indexes
CREATE INDEX idx_book_title ON book(title);
CREATE INDEX idx_author_first_name ON author(first_name);
CREATE INDEX idx_author_last_name ON author(last_name);

-- Indexes for join performance
CREATE INDEX idx_book_author_book ON book_author(book_id);
CREATE INDEX idx_book_author_author ON book_author(author_id);
CREATE INDEX idx_book_genre_book ON book_genre(book_id);
CREATE INDEX idx_book_genre_genre ON book_genre(genre_id);
CREATE INDEX idx_user_role_user ON user_role(user_id);
CREATE INDEX idx_user_role_role ON user_role(role_id);

-- Detailed book view
CREATE VIEW vw_books_detailed AS
SELECT
    b.id, b.title, b.publication_year, b.isbn,
    GROUP_CONCAT(DISTINCT CONCAT(a.first_name, ' ', a.last_name)) AS authors,
    GROUP_CONCAT(DISTINCT g.name) AS genres,
    b.created_at, b.updated_at
FROM book b
LEFT JOIN book_author ba on b.id = ba.book_id
LEFT JOIN author a ON ba.author_id = a.id
LEFT JOIN book_genre bg ON b.id = bg.book_id
LEFT JOIN genre g ON bg.genre_id = g.id
GROUP BY b.id;

-- Users detailed book view
CREATE OR REPLACE VIEW vw_user_books_detailed AS
SELECT
    ub.user_id, u.username, b.id AS book_id, b.title, b.publication_year,
    b.isbn, ub.status AS user_status, ub.notes, ub.added_at,
    GROUP_CONCAT(DISTINCT CONCAT(a.first_name, ' ', a.last_name)) AS authors,
    GROUP_CONCAT(DISTINCT g.name) AS genres, b.created_at, b.updated_at
FROM user_book ub
         JOIN user u ON ub.user_id = u.id
         JOIN book b ON ub.book_id = b.id
         LEFT JOIN book_author ba ON b.id = ba.book_id
         LEFT JOIN author a ON ba.author_id = a.id
         LEFT JOIN book_genre bg ON b.id = bg.book_id
         LEFT JOIN genre g ON bg.genre_id = g.id
GROUP BY ub.user_id, b.id;
