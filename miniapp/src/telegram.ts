import type { Telegram as TelegramType } from "telegram-web-app";

const Telegram = ((window as any).Telegram as TelegramType);
export default Telegram;
