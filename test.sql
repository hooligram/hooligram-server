DELETE FROM client;
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (1, 1, 1, 1);
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (2, 2, 2, 2);
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (3, 3, 3, 3);
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (4, 4, 4, 4);
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (5, 5, 5, 5);
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (6, 6, 6, 6);
INSERT INTO client (id, country_code, phone_number, verification_code) VALUES (7, 7, 7, 7);

DELETE FROM message_group;
INSERT INTO message_group (id, name) VALUES (1, "message group a (1, 2)");
INSERT INTO message_group (id, name) VALUES (2, "message group b (1, 2, 3)");
INSERT INTO message_group (id, name) VALUES (3, "message group c (3, 4, 5)");

DELETE FROM message_group_member;
INSERT INTO message_group_member (message_group_id, member_id) VALUES (1, 1);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (1, 2);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (2, 1);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (2, 2);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (2, 3);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (3, 3);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (3, 4);
INSERT INTO message_group_member (message_group_id, member_id) VALUES (3, 5);
