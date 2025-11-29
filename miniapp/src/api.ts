export const apiBase = import.meta.env.DEV ? '/api' : '';

export interface Settings {
    group_threshold: number;
}
