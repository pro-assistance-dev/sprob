ALTER TABLE passports
   add column adress varchar;

alter table document rename to documents;
alter table position rename to positions;

insert into positions  (id, name) values 
 ( 'ee246c4a-6939-400a-a669-cb8cce362ee9','Аналитик' ),
 ( '5e73bba5-c37c-4f14-aea8-a5ab0d5da073','Врач-педиатр' ),
 ( '66c16eb6-1e84-4bce-ba0a-040c142a7402','Мед. сестра постовая'),
 ( '423d25b2-0044-4a16-840e-f6cb299c4c77','Уборщица' );

insert into documents  (id, name) values 
 ( 'e7addf74-4c6b-4241-8c9d-18728d1c9f26','Справка о несудимости'),
 ( '0028f98e-bb42-4fae-a380-1fb1157c97ef','Справка из псих.диспансера'),
 ( '67fafb8a-6d65-4568-ae54-a8c3aae3036a','Справка из нарко-диспансера');

insert into positions_documents  (position_id, document_id) values 
 ( 'ee246c4a-6939-400a-a669-cb8cce362ee9','e7addf74-4c6b-4241-8c9d-18728d1c9f26'),
 ( 'ee246c4a-6939-400a-a669-cb8cce362ee9','0028f98e-bb42-4fae-a380-1fb1157c97ef'),
 ( '5e73bba5-c37c-4f14-aea8-a5ab0d5da073','e7addf74-4c6b-4241-8c9d-18728d1c9f26');
