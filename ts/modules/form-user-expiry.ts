export type InviteMode = "register" | "renew" | null;

export interface UserExpiryDisplayOptions {
    months: number;
    days: number;
    hours: number;
    minutes: number;
    now?: Date;
    formatDate: (date: Date) => string;
    userExpiryMessage: string;
    userExpiryRegisterMessage: string;
    userExpiryRenewalMessage: string;
}

export interface UserExpiryDisplay {
    visible: boolean;
    text: string;
}

export const userExpiryDisplayForInviteMode = (mode: InviteMode, options: UserExpiryDisplayOptions): UserExpiryDisplay => {
    if (mode == null) {
        return { visible: false, text: "" };
    }

    if (mode == "renew") {
        return {
            visible: options.userExpiryRenewalMessage != "",
            text: options.userExpiryRenewalMessage,
        };
    }

    const messageTemplate = options.userExpiryRegisterMessage || options.userExpiryMessage;
    if (messageTemplate == "") {
        return { visible: false, text: "" };
    }

    const time = options.now ? new Date(options.now.getTime()) : new Date();
    time.setMonth(time.getMonth() + options.months);
    time.setDate(time.getDate() + options.days);
    time.setHours(time.getHours() + options.hours);
    time.setMinutes(time.getMinutes() + options.minutes);

    return {
        visible: true,
        text: messageTemplate.replace("{date}", options.formatDate(time)),
    };
};
