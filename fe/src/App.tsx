
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  Bolt,
  Clock3,
  KeyRound,
  LogIn,
  LogOut,
  SendHorizonal,
  ShieldCheck,
  Signal,
  Smartphone,
  Trophy,
  UserRound,
  Wifi,
  WifiOff,
} from "lucide-react";
import { Toaster, toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

type ConnectionState = "idle" | "connecting" | "connected" | "disconnected";

type ApiEnvelope = {
  info?: string;
  data?: unknown;
};

type WsEnvelope = {
  type: string;
  data?: unknown;
};

type LeaderboardEntry = {
  username: string;
  rank: number;
  time: number;
};

type FeedItem = {
  id: number;
  text: string;
  at: string;
};

type ConnectOptions = {
  silent?: boolean;
  fromRestore?: boolean;
  sessionId?: string;
};

const PHONE_REGEX = /^1[3-9]\d{9}$/;
const USERNAME_REGEX = /^[a-zA-Z0-9_]{3,20}$/;
const CODE_REGEX = /^\d{6}$/;
const COUNTDOWN_TOTAL_MS = 60_000;
const SMS_COOLDOWN_SECONDS = 60;
const PRESS_LOCK_MS = 5_000;
const FEED_LIMIT = 6;
const SESSION_ID_STORAGE_KEY = "the-button.session_id";
const USERNAME_STORAGE_KEY = "the-button.username";
const GUEST_SESSION_ID = "guest";
const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080").replace(
  /\/+$/,
  "",
);

let feedSequence = 0;

function getStoredSessionId(): string {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(SESSION_ID_STORAGE_KEY)?.trim() ?? "";
}

function getStoredUsername(): string {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(USERNAME_STORAGE_KEY)?.trim() ?? "";
}

function setStoredSessionId(sessionId: string) {
  if (typeof window === "undefined") return;
  const normalized = sessionId.trim();
  if (!normalized) return;
  window.localStorage.setItem(SESSION_ID_STORAGE_KEY, normalized);
}

function setStoredUsername(username: string) {
  if (typeof window === "undefined") return;
  const normalized = username.trim();
  if (!normalized) return;
  window.localStorage.setItem(USERNAME_STORAGE_KEY, normalized);
}

function clearStoredAuth() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(SESSION_ID_STORAGE_KEY);
  window.localStorage.removeItem(USERNAME_STORAGE_KEY);
}

function toWsUrl(apiBaseUrl: string): string {
  const normalized = /^https?:\/\//.test(apiBaseUrl) ? apiBaseUrl : `http://${apiBaseUrl}`;
  try {
    const parsed = new URL(normalized);
    const wsProtocol = parsed.protocol === "https:" ? "wss:" : "ws:";
    return `${wsProtocol}//${parsed.host}/ws`;
  } catch {
    return "ws://localhost:8080/ws";
  }
}

function buildWsConnectUrl(baseWsUrl: string, sessionId: string): string {
  const normalized = sessionId.trim() || GUEST_SESSION_ID;
  try {
    const url = new URL(baseWsUrl);
    url.searchParams.set("session_id", normalized);
    return url.toString();
  } catch {
    return `${baseWsUrl}?session_id=${encodeURIComponent(normalized)}`;
  }
}

async function readEnvelope(response: Response): Promise<ApiEnvelope> {
  try {
    return (await response.json()) as ApiEnvelope;
  } catch {
    return {};
  }
}

function readInfo(envelope: ApiEnvelope, fallback: string): string {
  if (typeof envelope.info === "string" && envelope.info.trim().length > 0) return envelope.info;
  return fallback;
}

function readCaptchaData(data: unknown): { captchaId: string; imageBase64: string } | null {
  if (!data || typeof data !== "object") return null;
  const maybeData = data as { captcha_id?: unknown; image_base64?: unknown };
  if (typeof maybeData.captcha_id !== "string" || typeof maybeData.image_base64 !== "string") {
    return null;
  }
  return { captchaId: maybeData.captcha_id, imageBase64: maybeData.image_base64 };
}

function readLoginData(data: unknown): { sessionId: string; username: string } | null {
  if (!data || typeof data !== "object") return null;
  const maybeData = data as { session_id?: unknown; username?: unknown };
  if (typeof maybeData.session_id !== "string" || typeof maybeData.username !== "string") {
    return null;
  }
  const sessionId = maybeData.session_id.trim();
  const username = maybeData.username.trim();
  if (!sessionId || !username) return null;
  return { sessionId, username };
}

function parseWsEnvelope(raw: string): WsEnvelope | null {
  try {
    const payload = JSON.parse(raw) as WsEnvelope;
    if (typeof payload?.type !== "string") return null;
    return payload;
  } catch {
    return null;
  }
}

function readLeaderboard(data: unknown): LeaderboardEntry[] {
  if (!data || typeof data !== "object") return [];
  const list = (data as { entries?: unknown }).entries;
  if (!Array.isArray(list)) return [];

  return list
    .map((entry, index) => {
      if (!entry || typeof entry !== "object") return null;
      const maybeEntry = entry as { username?: unknown; rank?: unknown; time?: unknown };
      const username = typeof maybeEntry.username === "string" ? maybeEntry.username : "未知用户";
      const rank =
        typeof maybeEntry.rank === "number" && Number.isFinite(maybeEntry.rank)
          ? maybeEntry.rank
          : index + 1;
      const time =
        typeof maybeEntry.time === "number" && Number.isFinite(maybeEntry.time)
          ? maybeEntry.time
          : 0;
      return { username, rank, time };
    })
    .filter((entry): entry is LeaderboardEntry => entry !== null)
    .sort((a, b) => a.rank - b.rank);
}

function readStartTime(data: unknown): number | null {
  if (!data || typeof data !== "object") return null;
  const time = (data as { time?: unknown }).time;
  if (typeof time !== "number" || !Number.isFinite(time)) return null;
  return time;
}

function readPressUsername(data: unknown): string | null {
  if (!data || typeof data !== "object") return null;
  const username = (data as { username?: unknown }).username;
  if (typeof username !== "string" || username.trim().length === 0) return null;
  return username;
}

function formatCountdown(timeMs: number): string {
  const safe = Math.max(0, timeMs);
  const seconds = Math.floor(safe / 1000);
  const millis = Math.floor(safe % 1000).toString().padStart(3, "0");
  return `${seconds}.${millis}`;
}

function formatScoreTime(timeMs: number): string {
  return `${(timeMs / 1000).toFixed(3)}秒`;
}

function App() {
  const wsRef = useRef<WebSocket | null>(null);
  const restoreTriedRef = useRef(false);
  const initialSessionIdRef = useRef<string>(getStoredSessionId());
  const initialUsernameRef = useRef<string>(getStoredUsername());

  const [phone, setPhone] = useState("");
  const [username, setUsername] = useState("");
  const [captchaInput, setCaptchaInput] = useState("");
  const [smsCode, setSmsCode] = useState("");

  const [captchaId, setCaptchaId] = useState("");
  const [captchaImage, setCaptchaImage] = useState("");
  const [smsCooldown, setSmsCooldown] = useState(0);
  const [sendingSms, setSendingSms] = useState(false);
  const [loggingIn, setLoggingIn] = useState(false);
  const [loadingCaptcha, setLoadingCaptcha] = useState(false);
  const [loginModalOpen, setLoginModalOpen] = useState(false);

  const [connectionState, setConnectionState] = useState<ConnectionState>("idle");
  const [isLoggedIn, setIsLoggedIn] = useState<boolean>(initialSessionIdRef.current.length > 0);
  const [sessionId, setSessionId] = useState<string>(initialSessionIdRef.current);
  const [currentUser, setCurrentUser] = useState<string>(initialUsernameRef.current);
  const [serverAnchorMs, setServerAnchorMs] = useState<number | null>(null);
  const [remainingMs, setRemainingMs] = useState(COUNTDOWN_TOTAL_MS);
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [feed, setFeed] = useState<FeedItem[]>([]);
  const [pressCooldownUntil, setPressCooldownUntil] = useState(0);
  const [ticker, setTicker] = useState(Date.now());

  const wsUrl = useMemo(() => toWsUrl(API_BASE_URL), []);

  useEffect(() => {
    const timer = window.setInterval(() => setTicker(Date.now()), 100);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    if (smsCooldown <= 0) return;
    const timer = window.setInterval(() => {
      setSmsCooldown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [smsCooldown]);

  useEffect(() => {
    if (serverAnchorMs === null) {
      setRemainingMs(COUNTDOWN_TOTAL_MS);
      return;
    }
    const tick = () => {
      setRemainingMs(Math.max(0, COUNTDOWN_TOTAL_MS - (Date.now() - serverAnchorMs)));
    };
    tick();
    const timer = window.setInterval(tick, 40);
    return () => window.clearInterval(timer);
  }, [serverAnchorMs]);

  useEffect(() => {
    return () => {
      wsRef.current?.close();
    };
  }, []);

  const localLockMs = Math.max(0, pressCooldownUntil - ticker);
  const progressValue = (remainingMs / COUNTDOWN_TOTAL_MS) * 100;
  const topLeaderboard = leaderboard.slice(0, 20);

  const pushFeed = useCallback((text: string) => {
    const nextItem: FeedItem = {
      id: ++feedSequence,
      text,
      at: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
    };
    setFeed((prev) => [nextItem, ...prev].slice(0, FEED_LIMIT));
  }, []);

  const sendWs = useCallback((command: "1" | "2" | "3"): boolean => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return false;
    ws.send(command);
    return true;
  }, []);

  const requestSnapshot = useCallback(() => {
    sendWs("1");
    sendWs("2");
  }, [sendWs]);

  const loadCaptcha = useCallback(async () => {
    setLoadingCaptcha(true);
    try {
      const response = await fetch(`${API_BASE_URL}/sms/captcha`, {
        method: "GET",
        credentials: "include",
      });
      const payload = await readEnvelope(response);
      const captchaData = readCaptchaData(payload.data);
      if (!response.ok || !captchaData) {
        throw new Error(readInfo(payload, "获取图形验证码失败"));
      }
      setCaptchaId(captchaData.captchaId);
      setCaptchaImage(captchaData.imageBase64);
      setCaptchaInput("");
    } catch (error) {
      const message = error instanceof Error ? error.message : "获取图形验证码失败";
      toast.error(message);
    } finally {
      setLoadingCaptcha(false);
    }
  }, []);

  useEffect(() => {
    loadCaptcha();
  }, [loadCaptcha]);

  useEffect(() => {
    if (loginModalOpen) loadCaptcha();
  }, [loadCaptcha, loginModalOpen]);

  const handleWsMessage = useCallback(
    (raw: string) => {
      const envelope = parseWsEnvelope(raw);
      if (!envelope) return;

      switch (envelope.type) {
        case "time": {
          const startTime = readStartTime(envelope.data);
          if (startTime !== null) setServerAnchorMs(startTime);
          break;
        }
        case "leaderboard": {
          setLeaderboard(readLeaderboard(envelope.data));
          break;
        }
        case "button_press": {
          const actor = readPressUsername(envelope.data) ?? "未知用户";
          pushFeed(`${actor} 按下了按钮`);
          window.setTimeout(() => requestSnapshot(), 180);
          break;
        }
        case "lock": {
          setPressCooldownUntil(Date.now() + PRESS_LOCK_MS);
          toast.warning("你正处于 5 秒冷却中");
          break;
        }
        case "pending": {
          toast.info("本轮活动尚未开始");
          break;
        }
        case "finished": {
          toast.error("本轮活动已结束");
          break;
        }
        case "unauthorized": {
          setIsLoggedIn(false);
          setSessionId("");
          setCurrentUser("");
          clearStoredAuth();
          toast.error("请先登录后再按按钮");
          break;
        }
        default:
          break;
      }
    },
    [pushFeed, requestSnapshot],
  );

  const connectWs = useCallback(
    (options?: ConnectOptions) => {
      if (connectionState === "connecting") return;

      const silent = options?.silent ?? false;
      const fromRestore = options?.fromRestore ?? false;
      const connectSessionId =
        (options?.sessionId ?? (fromRestore ? initialSessionIdRef.current : sessionId)).trim() ||
        GUEST_SESSION_ID;
      let opened = false;

      wsRef.current?.close();
      setConnectionState("connecting");

      const ws = new WebSocket(buildWsConnectUrl(wsUrl, connectSessionId));
      wsRef.current = ws;

      ws.onopen = () => {
        opened = true;
        setConnectionState("connected");
        if (connectSessionId === GUEST_SESSION_ID) {
          setIsLoggedIn(false);
          setSessionId("");
          setCurrentUser("");
        } else {
          setIsLoggedIn(true);
          setSessionId(connectSessionId);
          setStoredSessionId(connectSessionId);
          if (currentUser.trim()) {
            setStoredUsername(currentUser);
          }
        }
        if (!silent) {
          pushFeed("已连接到游戏服务器");
          toast.success("连接成功");
        }
        requestSnapshot();
      };

      ws.onmessage = (event) => {
        handleWsMessage(String(event.data));
      };

      ws.onerror = () => {
        if (!silent) {
          toast.error("WebSocket 连接失败");
        }
      };

      ws.onclose = () => {
        if (wsRef.current === ws) {
          wsRef.current = null;
        }
        setConnectionState("disconnected");
        setPressCooldownUntil(0);

        if (fromRestore && !opened) {
          setIsLoggedIn(false);
          setSessionId("");
          setCurrentUser("");
          clearStoredAuth();
          return;
        }

        if (!silent) {
          pushFeed("连接已关闭");
        }
      };
    },
    [connectionState, currentUser, handleWsMessage, pushFeed, requestSnapshot, sessionId, wsUrl],
  );

  useEffect(() => {
    if (restoreTriedRef.current) return;
    restoreTriedRef.current = true;
    connectWs({
      silent: true,
      fromRestore: Boolean(initialSessionIdRef.current),
    });
  }, [connectWs]);

  const disconnectWs = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    setConnectionState("disconnected");
    setPressCooldownUntil(0);
  }, []);

  const handleLogout = useCallback(() => {
    disconnectWs();
    setIsLoggedIn(false);
    setSessionId("");
    setCurrentUser("");
    setPhone("");
    setUsername("");
    setSmsCode("");
    setCaptchaInput("");
    setLoginModalOpen(false);
    clearStoredAuth();
    toast.success("已退出登录");
    connectWs({ silent: true, sessionId: GUEST_SESSION_ID });
  }, [connectWs, disconnectWs]);

  const handleSendSms = useCallback(async () => {
    const normalizedPhone = phone.trim();
    const normalizedCaptcha = captchaInput.trim();

    if (!PHONE_REGEX.test(normalizedPhone)) {
      toast.error("手机号格式不正确");
      return;
    }
    if (!captchaId || normalizedCaptcha.length === 0) {
      toast.error("请输入图形验证码");
      return;
    }
    if (smsCooldown > 0 || sendingSms) return;

    setSendingSms(true);
    try {
      const response = await fetch(`${API_BASE_URL}/sms/code`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          phone_number: normalizedPhone,
          captcha_id: captchaId,
          code: normalizedCaptcha,
        }),
      });

      const payload = await readEnvelope(response);
      if (!response.ok) {
        throw new Error(readInfo(payload, "发送短信验证码失败"));
      }
      toast.success(readInfo(payload, "短信验证码已发送"));
      setSmsCooldown(SMS_COOLDOWN_SECONDS);
      setSmsCode("");
    } catch (error) {
      const message = error instanceof Error ? error.message : "发送短信验证码失败";
      toast.error(message);
      loadCaptcha();
    } finally {
      setSendingSms(false);
    }
  }, [captchaId, captchaInput, loadCaptcha, phone, sendingSms, smsCooldown]);

  const handleLogin = useCallback(async () => {
    const normalizedPhone = phone.trim();
    const normalizedUsername = username.trim();
    const normalizedCode = smsCode.trim();

    if (!PHONE_REGEX.test(normalizedPhone)) {
      toast.error("手机号格式不正确");
      return;
    }
    if (!USERNAME_REGEX.test(normalizedUsername)) {
      toast.error("用户名需 3-20 位，仅支持字母/数字/下划线");
      return;
    }
    if (!CODE_REGEX.test(normalizedCode)) {
      toast.error("短信验证码必须为 6 位数字");
      return;
    }

    setLoggingIn(true);
    try {
      const response = await fetch(`${API_BASE_URL}/sms/verify`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: normalizedUsername,
          phone_number: normalizedPhone,
          verify_code: normalizedCode,
        }),
      });

      const payload = await readEnvelope(response);
      if (!response.ok) {
        throw new Error(readInfo(payload, "登录失败"));
      }
      const loginData = readLoginData(payload.data);
      if (!loginData) {
        throw new Error("登录返回数据异常");
      }

      setSessionId(loginData.sessionId);
      setStoredSessionId(loginData.sessionId);
      setCurrentUser(loginData.username);
      setStoredUsername(loginData.username);
      setIsLoggedIn(true);
      setLoginModalOpen(false);
      toast.success(readInfo(payload, "登录成功"));
      connectWs({ sessionId: loginData.sessionId });
    } catch (error) {
      const message = error instanceof Error ? error.message : "登录失败";
      toast.error(message);
    } finally {
      setLoggingIn(false);
    }
  }, [connectWs, phone, smsCode, username]);

  const handlePressButton = useCallback(() => {
    if (connectionState !== "connected") {
      toast.error("请先连接 WebSocket");
      return;
    }
    if (!isLoggedIn) {
      toast.error("请先登录后再按按钮");
      return;
    }
    if (remainingMs <= 0) {
      toast.error("本轮活动已结束");
      return;
    }
    if (localLockMs > 0) {
      toast.warning("冷却中，请稍后");
      return;
    }
    if (!sendWs("3")) {
      toast.error("发送失败");
      return;
    }
    setPressCooldownUntil(Date.now() + PRESS_LOCK_MS);
  }, [connectionState, isLoggedIn, localLockMs, remainingMs, sendWs]);

  const connectionInfo =
    connectionState === "connected"
      ? "已连接"
      : connectionState === "connecting"
        ? "连接中..."
        : connectionState === "disconnected"
          ? "已断开"
          : "未连接";

  const canPress =
    connectionState === "connected" && isLoggedIn && remainingMs > 0 && localLockMs <= 0;

  return (
    <>
      <div className="min-h-screen bg-[radial-gradient(1200px_500px_at_20%_0%,rgba(14,148,161,.12),transparent_70%),radial-gradient(1000px_500px_at_95%_95%,rgba(240,124,67,.14),transparent_65%)] pb-10">
        <main className="container mx-auto px-4 pt-6 md:px-6 md:pt-8">
          <section className="mb-6 flex flex-col gap-4 md:mb-8 md:flex-row md:items-start md:justify-between">
            <div className="animate-in fade-in slide-in-from-top-4 duration-500">
              <p className="mb-2 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-semibold uppercase tracking-widest text-primary">
                <Signal className="h-3.5 w-3.5" />
                按钮挑战
              </p>
              <h1 className="text-3xl font-bold tracking-tight text-foreground md:text-4xl">
                实时反应竞技场
              </h1>
              <p className="mt-2 max-w-2xl text-sm text-muted-foreground md:text-base">
                先用短信完成登录，再进入 60 秒循环挑战。按下时机越精准，排行榜成绩越高。
              </p>
            </div>

            <div className="flex w-full flex-col items-start gap-3 md:w-auto md:items-end">
              <div className="flex items-center gap-2">
                {isLoggedIn ? (
                  <>
                    <div className="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-white/80 px-3 py-1.5 text-sm font-medium shadow-sm backdrop-blur">
                      <UserRound className="h-4 w-4 text-primary" />
                      <span>{currentUser}</span>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      className="gap-2 rounded-full px-3"
                      onClick={handleLogout}
                    >
                      <LogOut className="h-4 w-4" />
                      退出
                    </Button>
                  </>
                ) : (
                  <Button
                    type="button"
                    className="gap-2 rounded-full px-4"
                    onClick={() => setLoginModalOpen(true)}
                  >
                    <LogIn className="h-4 w-4" />
                    登录
                  </Button>
                )}
              </div>
              <div className="flex flex-wrap gap-2">
                <Badge variant="secondary" className="rounded-full px-3 py-1">
                  API {API_BASE_URL}
                </Badge>
                <Badge variant="outline" className="rounded-full px-3 py-1">
                  WS {wsUrl}
                </Badge>
              </div>
            </div>
          </section>

          <section className="space-y-6">
            <Card className="animate-in fade-in slide-in-from-bottom-6 duration-700 border-white/60 bg-white/90 backdrop-blur">
              <CardHeader className="pb-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Bolt className="h-5 w-5 text-accent" />
                      游戏面板
                    </CardTitle>
                    <CardDescription>
                      每 5 秒可按一次按钮，间隔越小，排行榜成绩越高。
                    </CardDescription>
                  </div>
                  <Badge
                    variant={connectionState === "connected" ? "default" : "outline"}
                    className="rounded-full px-3 py-1"
                  >
                    {connectionState === "connected" ? (
                      <Wifi className="mr-1 h-3.5 w-3.5" />
                    ) : (
                      <WifiOff className="mr-1 h-3.5 w-3.5" />
                    )}
                    {connectionInfo}
                  </Badge>
                </div>
              </CardHeader>

              <CardContent className="space-y-6">
                <div className="grid gap-4 rounded-xl border border-border/70 bg-muted/40 p-4 sm:grid-cols-3">
                  <div>
                    <p className="text-xs uppercase tracking-wide text-muted-foreground">剩余时间</p>
                    <p className="mono-digits mt-1 text-2xl font-bold">{formatCountdown(remainingMs)}</p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-wide text-muted-foreground">冷却时间</p>
                    <p className="mono-digits mt-1 text-2xl font-bold">{(localLockMs / 1000).toFixed(1)}秒</p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-wide text-muted-foreground">会话状态</p>
                    <p className="mt-1 text-2xl font-bold">{isLoggedIn ? currentUser : "游客"}</p>
                  </div>
                  <div className="sm:col-span-3">
                    <Progress value={progressValue} className="h-2.5" />
                  </div>
                </div>

                <div className="flex justify-center">
                  <Button
                    type="button"
                    onClick={handlePressButton}
                    disabled={!canPress}
                    className={cn(
                      "relative h-52 w-52 rounded-full border border-white/70 text-xl font-bold tracking-wide shadow-[0_14px_35px_rgba(0,0,0,0.15)] transition duration-200 md:h-60 md:w-60 md:text-2xl",
                      "bg-[linear-gradient(150deg,hsl(22,88%,58%),hsl(12,86%,52%))] text-white hover:scale-[1.03] hover:brightness-105",
                      "disabled:scale-100 disabled:bg-muted disabled:text-muted-foreground disabled:shadow-none",
                    )}
                  >
                    {connectionState !== "connected"
                      ? "离线"
                      : remainingMs <= 0
                        ? "已结束"
                        : localLockMs > 0
                          ? `${(localLockMs / 1000).toFixed(1)}秒`
                          : "按下"}
                  </Button>
                </div>
              </CardContent>
            </Card>

            <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
              <Card className="border-white/60 bg-white/90 backdrop-blur">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-base">
                    <Clock3 className="h-4 w-4 text-primary" />
                    实时动态
                  </CardTitle>
                  <CardDescription>房间内最新事件流。</CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                  {feed.length === 0 ? (
                    <p className="text-sm text-muted-foreground">暂无动态，连接后即可开始游戏。</p>
                  ) : (
                    feed.map((item) => (
                      <div
                        key={item.id}
                        className="flex items-center justify-between rounded-lg border border-border/60 bg-muted/40 px-3 py-2 text-sm"
                      >
                        <span>{item.text}</span>
                        <span className="mono-digits text-xs text-muted-foreground">{item.at}</span>
                      </div>
                    ))
                  )}
                </CardContent>
              </Card>

              <Card className="border-white/60 bg-white/90 backdrop-blur">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-base">
                    <Trophy className="h-4 w-4 text-accent" />
                    排行榜
                  </CardTitle>
                  <CardDescription>前 20 名（成绩越高越靠前）。</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="rounded-lg border border-border/70">
                    <div className="grid grid-cols-[70px_1fr_120px] items-center bg-muted/55 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                      <span>排名</span>
                      <span>玩家</span>
                      <span className="text-right">间隔</span>
                    </div>
                    <Separator />
                    <div className="max-h-[360px] overflow-auto">
                      {topLeaderboard.length === 0 ? (
                        <p className="px-3 py-6 text-center text-sm text-muted-foreground">暂无排行榜数据。</p>
                      ) : (
                        topLeaderboard.map((entry) => (
                          <div
                            key={`${entry.rank}-${entry.username}`}
                            className="grid grid-cols-[70px_1fr_120px] items-center px-3 py-2 text-sm odd:bg-white even:bg-muted/30"
                          >
                            <span className="mono-digits font-semibold">#{entry.rank}</span>
                            <span className="truncate pr-2">{entry.username}</span>
                            <span className="mono-digits text-right font-medium">{formatScoreTime(entry.time)}</span>
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </section>
        </main>
      </div>

      <Dialog open={loginModalOpen} onOpenChange={setLoginModalOpen}>
        <DialogContent className="sm:max-w-md border-white/70 bg-white/95 backdrop-blur">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <ShieldCheck className="h-5 w-5 text-primary" />
              登录验证
            </DialogTitle>
            <DialogDescription>
              完成短信验证后会获取 `session_id`，随后通过 ws query 参数进行身份标识。
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="phone">手机号</Label>
              <div className="relative">
                <Smartphone className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  id="phone"
                  inputMode="numeric"
                  placeholder="请输入 11 位手机号"
                  className="pl-9"
                  value={phone}
                  onChange={(event) => setPhone(event.target.value)}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <div className="relative">
                <UserRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  id="username"
                  placeholder="3-20 位，字母/数字/下划线"
                  className="pl-9"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="captcha">图形验证码</Label>
              <div className="flex gap-2">
                <Input
                  id="captcha"
                  placeholder="请输入图形验证码"
                  value={captchaInput}
                  onChange={(event) => setCaptchaInput(event.target.value)}
                />
                <button
                  type="button"
                  onClick={loadCaptcha}
                  className="overflow-hidden rounded-md border bg-muted"
                  disabled={loadingCaptcha}
                  title="刷新验证码"
                >
                  {captchaImage ? (
                    <img
                      src={captchaImage}
                      alt="图形验证码"
                      className={cn("h-10 w-28 object-cover", loadingCaptcha && "opacity-60")}
                    />
                  ) : (
                    <div className="flex h-10 w-28 items-center justify-center text-xs text-muted-foreground">
                      加载中
                    </div>
                  )}
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="sms">短信验证码</Label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="sms"
                    placeholder="请输入 6 位验证码"
                    inputMode="numeric"
                    className="pl-9"
                    value={smsCode}
                    onChange={(event) => setSmsCode(event.target.value)}
                  />
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  className="min-w-28"
                  onClick={handleSendSms}
                  disabled={sendingSms || smsCooldown > 0}
                >
                  {smsCooldown > 0 ? `${smsCooldown}秒` : sendingSms ? "发送中..." : "发送验证码"}
                </Button>
              </div>
            </div>

            <Button
              type="button"
              className="w-full gap-2"
              onClick={handleLogin}
              disabled={loggingIn}
            >
              <SendHorizonal className="h-4 w-4" />
              {loggingIn ? "验证中..." : "登录 / 注册"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Toaster richColors position="top-right" />
    </>
  );
}

export default App;
