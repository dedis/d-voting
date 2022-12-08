export interface ConfigInterface {
    getString(key: string): string;
    getStrings(key: string): string[];
    getBool(key: string): boolean;
    getInt(key: string): number;
    getFloat(key: string): number;
    set(key: string, value: string): void;
}
export declare class Config implements ConfigInterface {
    private static DEFAULT_SECTION;
    private static DEFAULT_COMMENT;
    private static DEFAULT_COMMENT_SEM;
    private static DEFAULT_MULTI_LINE_SEPARATOR;
    private data;
    private constructor();
    /**
     * newConfig create an empty configuration representation from file.
     *
     * @param confName the path of the model file.
     * @return the constructor of Config.
     */
    static newConfig(confName: string): Config;
    /**
     * newConfigFromText create an empty configuration representation from text.
     *
     * @param text the model text.
     * @return the constructor of Config.
     */
    static newConfigFromText(text: string): Config;
    /**
     * addConfig adds a new section->key:value to the configuration.
     */
    private addConfig;
    private parse;
    private parseBuffer;
    private write;
    getBool(key: string): boolean;
    getInt(key: string): number;
    getFloat(key: string): number;
    getString(key: string): string;
    getStrings(key: string): string[];
    set(key: string, value: string): void;
    get(key: string): string;
}
