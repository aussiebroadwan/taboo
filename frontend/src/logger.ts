interface LogFields {
    [key: string]: unknown;
}

interface Logger {
    debug(msg: string, fields?: LogFields): void;
    info(msg: string, fields?: LogFields): void;
    warn(msg: string, fields?: LogFields): void;
    error(msg: string, fields?: LogFields): void;
    with(fields: LogFields): Logger;
}

function createLogger(defaultFields: LogFields = {}): Logger {
    function formatArgs(msg: string, fields?: LogFields): [string] | [string, LogFields] {
        const merged = { ...defaultFields, ...fields };
        if (Object.keys(merged).length === 0) {
            return [msg];
        }
        return [msg, merged];
    }

    return {
        debug(msg: string, fields?: LogFields) { console.debug(...formatArgs(msg, fields)); },
        info(msg: string, fields?: LogFields) { console.info(...formatArgs(msg, fields)); },
        warn(msg: string, fields?: LogFields) { console.warn(...formatArgs(msg, fields)); },
        error(msg: string, fields?: LogFields) { console.error(...formatArgs(msg, fields)); },
        with(fields: LogFields) { return createLogger({ ...defaultFields, ...fields }); },
    };
}

const logger = createLogger();
export default logger;
export type { Logger, LogFields };
