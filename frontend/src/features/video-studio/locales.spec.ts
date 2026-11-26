import { baseCompile } from '@intlify/message-compiler';
import { expect, it } from 'vitest';
import en from '@/i18n/locales/en/videoStudio';
import zh from '@/i18n/locales/zh/videoStudio';
it('compiles every video message, including the literal reference mention character', () => {
    for (const messages of [en, zh]) {
        for (const message of Object.values(messages.videoStudio)) {
            const errors: unknown[] = [];
            baseCompile(message, { onError: error => errors.push(error) });
            expect(errors, message).toEqual([]);
        }
    }
});
