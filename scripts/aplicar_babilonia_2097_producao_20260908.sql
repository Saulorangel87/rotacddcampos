-- Correção manual da Rua Babilônia - ID 2097
BEGIN;
UPDATE ruas SET geometria = '{"type":"MultiLineString","coordinates":[[[-41.31667,-21.5996855],[-41.3157013,-21.599811],[-41.3134548,-21.6000719]]]}', updated_at = now() WHERE id = 2097 AND ativo = true;
DO $$ DECLARE total integer; BEGIN SELECT count(*) INTO total FROM ruas WHERE ativo = true AND id = 2097; IF total <> 1 THEN RAISE EXCEPTION 'ID 2097 não está ativo após atualização'; END IF; END $$;
COMMIT;
