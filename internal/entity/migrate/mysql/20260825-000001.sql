UPDATE albums
SET album_order = 'oldest'
WHERE album_type = 'folder'
  AND (album_order = 'added' OR album_order = '' OR album_order IS NULL);
