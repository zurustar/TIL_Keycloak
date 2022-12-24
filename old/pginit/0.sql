

DROP TABLE IF EXISTS TagTypes;
DROP TABLE IF EXISTS Tags;

CREATE TABLE TagTypes (
    tag_type_id INTEGER
    , name TEXT
);
INSERT INTO TagTypes(tag_type_id, name) VALUES(1, '会社')
INSERT INTO TagTypes(tag_type_id, name) VALUES(2, '製品')

CREATE TABLE Tags (
    tag_id INTEGER
    , tag_type_id INTEGER
    , name TEXT
    , FOREIGN KEY (tag_type_id) REFERENCES TagTypes(tag_type_id)
);

INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(1, 1, 'トヨタ');
INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(2, 1, 'ホンダ');
INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(3, 1, '日産');
INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(4, 1, 'マツダ');
INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(5, 1, 'スバル');
INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(6, 1, 'スズキ');
INSERT INTO Tags(tag_id, tag_type_id, name) VALUES(7, 1, 'ダイハツ');
