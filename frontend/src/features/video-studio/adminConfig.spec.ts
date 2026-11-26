import { describe, it, expect } from 'vitest';
import { mergeVideoModels, setVideoVisibility } from './adminConfig';
const catalog = [{ id: '1:a', group_id: 1, group_name: 'G', model: 'a', label: 'A', protocol: 'grok' as const }, { id: '2:b', group_id: 2, group_name: 'H', model: 'b', label: 'B', protocol: 'grok' as const }];
describe('video admin configuration', () => {
    it('keeps retired overrides and changes only explicitly filtered valid rows', () => {
        const settings = [{ group_id: 99, model: 'old', label: 'Old', enabled: true }, { group_id: 2, model: 'b', label: 'Custom', enabled: true }];
        const rows = mergeVideoModels(catalog, settings);
        expect(rows).toHaveLength(3);
        const changed = setVideoVisibility(settings, rows.filter(r => r.group_id === 1), false, [1, 2]);
        expect(changed).toEqual([...settings, { group_id: 1, model: 'a', label: 'A', protocol: 'grok', enabled: false }]);
        expect(setVideoVisibility(settings, rows.filter(r => r.group_id === 99), false, [1, 2])).toEqual(settings);
    });
});
