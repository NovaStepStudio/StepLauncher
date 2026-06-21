export class JvmArgsHelper {
    static merge(globalArgs: string[], instanceArgs: string[]): string[] {
        const merged = new Map<string, string>()

        const processArg = (arg: string) => {
            if (arg.startsWith("-Xmx")) {
                merged.set("-Xmx", arg)
            } else if (arg.startsWith("-Xms")) {
                merged.set("-Xms", arg)
            } else if (arg.startsWith("-D")) {
                const eq = arg.indexOf("=")
                const key = eq > 0 ? arg.substring(0, eq) : arg
                merged.set(key, arg)
            } else {
                merged.set(arg, arg)
            }
        }

        globalArgs.forEach(processArg)
        instanceArgs.forEach(processArg)

        return Array.from(merged.values())
    }

    static clean(args: string[]): string[] {
        return args.filter(a => a && a.trim().length > 0)
    }

    static parseString(input: string): string[] {
        const regex = /[^\s"']+|"([^"]*)"|'([^']*)'/g
        const args: string[] = []
        let match

        while ((match = regex.exec(input)) !== null) {
            if (match[1]) args.push(match[1])
            else if (match[2]) args.push(match[2])
            else args.push(match[0])
        }

        return args
    }
}
