CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS users CASCADE;
CREATE UNLOGGED TABLE users
(
    ID       SERIAL NOT NULL PRIMARY KEY,
    nickname CITEXT NOT NULL UNIQUE COLLATE "POSIX",
    fullname TEXT   NOT NULL,
    email    CITEXT NOT NULL UNIQUE,
    about    TEXT
);
--indexes
CREATE INDEX idx_nick_nick ON users (nickname);
CREATE INDEX idx_nick_email ON users (email);
CREATE INDEX idx_nick_cover ON users (nickname, fullname, about, email);

DROP TABLE IF EXISTS forums CASCADE;
CREATE UNLOGGED TABLE forums
(
    ID        SERIAL                             NOT NULL PRIMARY KEY,

    slug      CITEXT                             NOT NULL UNIQUE,
    threads   INTEGER DEFAULT 0                  NOT NULL,
    posts     INTEGER DEFAULT 0                  NOT NULL,
    title     TEXT                               NOT NULL,
    user_nick CITEXT REFERENCES users (nickname) NOT NULL
);
--indexes
CREATE INDEX idx_forum_slug ON forums using hash(slug);

DROP TABLE IF EXISTS threads CASCADE;
CREATE UNLOGGED TABLE threads
(
    ID      SERIAL                          NOT NULL PRIMARY KEY,
    author  CITEXT                          NOT NULL REFERENCES users (nickname),
    created TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    forum   CITEXT REFERENCES forums (slug) NOT NULL,
    message TEXT                            NOT NULL,
    slug    CITEXT UNIQUE,
    title   TEXT                            NOT NULL,
    votes   INTEGER DEFAULT 0
);
--indexes
CREATE INDEX idx_thread_id ON threads(id);
CREATE INDEX idx_thread_slug ON threads(slug);
CREATE INDEX idx_thread_coverage ON threads (forum, created, id, slug, author, title, message, votes);

DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS forum_users;
/*CREATE TABLE forum_users
(
    forumID INTEGER REFERENCES forums (ID),
    userID  INTEGER REFERENCES users (ID)
);
*/
CREATE UNLOGGED TABLE forum_users
(
    forum citext REFERENCES forums (slug),
    nickname  citext REFERENCES users (nickname)
);

ALTER TABLE IF EXISTS forum_users ADD CONSTRAINT uniq UNIQUE (forum, nickname);
CREATE INDEX idx_forum_user ON forum_users (forum, nickname);
CREATE INDEX users_forum_forum_index ON forum_users (forum); -- +
CREATE INDEX users_forum_user_index ON forum_users (nickname);

/*ALTER TABLE IF EXISTS forum_users ADD CONSTRAINT uniq UNIQUE (forumID, userID);
CREATE INDEX idx_forum_user ON forum_users (forumID, userID);*/

DROP TABLE IF EXISTS votes;
CREATE UNLOGGED TABLE votes
(
    user_nick CITEXT REFERENCES users (nickname) NOT NULL,
    voice BOOLEAN NOT NULL,
    thread  INTEGER REFERENCES threads (ID) NOT NULL
);
ALTER TABLE IF EXISTS votes ADD CONSTRAINT uniq_votes UNIQUE (user_nick, thread);
CREATE INDEX idx_vote ON votes(thread, voice);


-----------------------------------------------------
/*CREATE TABLE posts
(
    id      integer                                NOT NULL PRIMARY KEY,
    author  citext                                 NOT NULL REFERENCES users (nickname),
    created text                                   NOT NULL,

    forum   CITEXT REFERENCES forums (slug)        NOT NULL,

    edited  boolean DEFAULT false                  NOT NULL,
    message text                                   NOT NULL,
    parent  integer DEFAULT 0                      NOT NULL,
    thread  INTEGER REFERENCES public.threads (ID) NOT NULL,
    path    INTEGER[] DEFAULT '{0}':: INTEGER [] NOT NULL
);

CREATE SEQUENCE if not exists post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE post_id_seq OWNED BY posts.id;
ALTER TABLE ONLY posts ALTER COLUMN id SET DEFAULT nextval('post_id_seq'::regclass);
SELECT pg_catalog.setval('post_id_seq', 1, false);
*/

CREATE UNLOGGED TABLE posts
(
    id      serial                                 PRIMARY KEY,
    author  citext                                 NOT NULL,
    created text                                   ,

    forum   CITEXT                                  ,

    edited  boolean DEFAULT false                  ,
    message text                                   NOT NULL,
    parent  integer DEFAULT 0                      ,
    thread  INTEGER ,
    path    INTEGER[] DEFAULT '{0}':: INTEGER [] ,
    FOREIGN KEY (author) REFERENCES "users" (nickname),
    FOREIGN KEY (forum) REFERENCES "forums" (slug),
    FOREIGN KEY (thread) REFERENCES "threads" (id)
 --   FOREIGN KEY (parent) REFERENCES "posts" (id)
);

/*CREATE INDEX post_author_forum_index ON posts USING btree (author, forum);
CREATE INDEX post_forum_index ON posts USING btree (forum);
CREATE INDEX post_parent_index ON posts USING btree (parent);
CREATE INDEX post_path_index ON posts USING gin (path);
CREATE INDEX post_thread_index ON posts USING btree (thread);*/


CREATE OR REPLACE FUNCTION update_path() RETURNS TRIGGER AS
$update_path$
DECLARE
parent_path         INT[];
    first_parent_thread INT;
BEGIN
   -- IF (NEW.parent IS NULL) THEN
 IF (NEW.parent = 0) THEN
     new.path = array[new.id::INTEGER];
--        NEW.path := array_append(new.path, new.id);
ELSE
SELECT path FROM posts WHERE id = new.parent INTO parent_path;
SELECT thread FROM posts WHERE id = parent_path[1] INTO first_parent_thread;
IF NOT FOUND OR first_parent_thread != NEW.thread THEN
            RAISE EXCEPTION 'parent is from different thread' USING ERRCODE = '00409';
end if;
            NEW.path := parent_path || new.id;
       -- NEW.path := NEW.path || parent_path || new.id;
end if;
UPDATE forums SET posts=posts + 1 WHERE lower(forums.slug) = lower(new.forum);
RETURN new;
end
$update_path$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_user_forum() RETURNS TRIGGER AS
$update_users_forum$
BEGIN
INSERT INTO forum_users (nickname, forum) VALUES (NEW.author, NEW.forum) on conflict do nothing;
return NEW;
end
$update_users_forum$ LANGUAGE plpgsql;

CREATE TRIGGER post_insert_user_forum
    AFTER INSERT
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_user_forum();

CREATE TRIGGER path_update_trigger
    BEFORE INSERT
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_path();

CREATE INDEX post_first_parent_thread_index ON posts ((posts.path[1]), thread);
CREATE INDEX post_first_parent_id_index ON posts ((posts.path[1]), id);
CREATE INDEX post_first_parent_index ON posts ((posts.path[1]));
CREATE INDEX post_path_index ON posts ((posts.path));
CREATE INDEX post_thread_index ON posts (thread); -- -
CREATE INDEX post_thread_id_index ON posts (thread, id); -- +

CREATE INDEX post_path_id_index ON posts (id, (posts.path));
CREATE INDEX post_thread_path_id_index ON posts (thread, (posts.parent), id);
CREATE INDEX forum_slug_lower_index ON forums (lower(forums.slug)); -- +
