// Fake backend to work on the web without the microservices: `npm run mock` and, in another
// terminal, `npm run dev:mock` (Vite with .env.mock, which proxies everything here). Everything lives in memory and is lost on restart.
//
// It answers the routes the app uses with the same shapes as the real services (auth, users,
// books, uploads, download, websocket with the "bearer" subprotocol) and simulates the pipeline:
// after PATCH /uploads/done each source walks processing → (joined | kepubifying) → done → sent,
// with websocket events, join results for "should_join" folders and the close of the folder.
// Any email and password sign in; the account "error@inkomi.dev" fails its uploads.
import http from 'node:http';
import { randomUUID } from 'node:crypto';
import { WebSocketServer } from 'ws';

const port = Number(process.env.MOCK_PORT ?? 8080);
const origin = `http://localhost:${port}`;
const stepMs = Number(process.env.MOCK_STEP_MS ?? 1500);
const accessTtlMs = 15 * 60_000;

// ─── state ───────────────────────────────────────────────────────────────────────────────

/** access token → user id; refresh token → user id (single use, like the real one). */
const accessTokens = new Map();
const refreshTokens = new Map();
/** user id → profile */
const users = new Map();
/** source id → row (camelCase, like GET /uploads) plus userId */
const sources = new Map();
/** user id → websocket (the server keeps one per user) */
const sockets = new Map();

const now = () => new Date().toISOString();
const minutesAgo = (m) => new Date(Date.now() - m * 60_000).toISOString();

function userFor(email) {
  for (const user of users.values()) if (user.email === email) return user;
  const user = {
    id: users.size + 1,
    email,
    emailVerified: true,
    hasPassword: true,
    providers: [],
    dropbox: { linked: false, folder: '/Apps/Inkomi' },
  };
  users.set(user.id, user);
  seedHistory(user.id);
  return user;
}

function issueTokens(userId) {
  const accessToken = `mock-access-${randomUUID()}`;
  const refreshToken = `mock-refresh-${randomUUID()}`;
  accessTokens.set(accessToken, userId);
  refreshTokens.set(refreshToken, userId);
  return {
    accessToken,
    accessTokenExpiresAt: new Date(Date.now() + accessTtlMs).toISOString(),
    refreshToken,
  };
}

function row(userId, fields) {
  const created = fields.createdAt ?? now();
  return {
    id: randomUUID(),
    title: '',
    filename: '',
    kepubify: false,
    toCloud: false,
    isRtl: false,
    join: false,
    status: 'pending_upload',
    errorMsg: null,
    size: null,
    isFile: true,
    createdAt: created,
    updatedAt: created,
    completedAt: null,
    parentId: null,
    position: 0,
    isResult: false,
    userId,
    ...fields,
  };
}

function seedHistory(userId) {
  const add = (fields) => {
    const r = row(userId, fields);
    sources.set(r.id, r);
    return r;
  };
  add({
    title: 'El Quijote',
    filename: 'El Quijote.epub',
    status: 'sent',
    size: 2_100_000,
    createdAt: minutesAgo(90),
    completedAt: minutesAgo(85),
  });
  add({
    title: 'Akira',
    filename: 'Akira.pdf',
    status: 'failed',
    errorMsg: 'archivo corrupto',
    createdAt: minutesAgo(60 * 26),
    completedAt: minutesAgo(60 * 26),
  });
  const folder = add({
    title: 'Berserk 1-3',
    filename: 'Berserk 1-3',
    isFile: false,
    join: true,
    status: 'done',
    createdAt: minutesAgo(60 * 24 * 3),
    completedAt: minutesAgo(60 * 24 * 3),
  });
  [1, 2, 3].forEach((n) =>
    add({
      title: `Berserk ${n}`,
      filename: `Berserk ${n}.cbz`,
      parentId: folder.id,
      position: n - 1,
      status: 'joined',
      createdAt: folder.createdAt,
    })
  );
  add({
    title: 'Berserk 1-3',
    filename: 'Berserk 1-3.epub',
    parentId: folder.id,
    isResult: true,
    status: 'sent',
    size: 340_000_000,
    createdAt: folder.createdAt,
    completedAt: folder.completedAt,
  });
}

const books = [
  ['Don Quijote de la Mancha', 'Miguel de Cervantes', 'epub', 1080],
  ['El Quijote (ilustrado)', 'Miguel de Cervantes', 'pdf', 620],
  ['Quijote 1 (manga)', 'Varios', 'cbz', 190],
  ['El ingenioso hidalgo', 'Miguel de Cervantes', 'epub', 940],
  ['Dune', 'Frank Herbert', 'epub', 720],
  ['Berserk 1', 'Kentaro Miura', 'cbz', 224],
  ['One Piece 1', 'Eiichiro Oda', 'cbz', 208],
  ['Akira 1', 'Katsuhiro Otomo', 'pdf', 360],
].map(([title, author, extension, pages], i) => ({
  title,
  author,
  pages,
  language: 'Español',
  extension,
  // The last one has no md5, like some mirror results: it can't be added.
  md5: i === 7 ? '' : i.toString(16).repeat(32),
  coverUrl: '',
}));

// ─── pipeline ────────────────────────────────────────────────────────────────────────────

const terminal = (status) => ['done', 'sent', 'failed'].includes(status);
const childrenOf = (id) => [...sources.values()].filter((s) => s.parentId === id);

function send(userId, event) {
  const ws = sockets.get(userId);
  if (ws?.readyState === 1) ws.send(JSON.stringify(event));
}

function setStatus(source, status, extra = {}) {
  Object.assign(source, { status, updatedAt: now(), ...extra });
  if (terminal(status)) source.completedAt ??= now();
  const { userId, ...rest } = source;
  send(userId, {
    id: source.id,
    status,
    filename: rest.filename,
    size: rest.size ?? undefined,
    error: rest.errorMsg ?? undefined,
  });
}

const delay = () => new Promise((r) => setTimeout(r, stepMs * (0.6 + Math.random() * 0.8)));

async function runSource(source) {
  const user = users.get(source.userId);
  const parent = source.parentId ? sources.get(source.parentId) : null;
  const isEpub = source.filename.toLowerCase().endsWith('.epub');

  await delay();
  setStatus(source, 'processing');
  await delay();
  if (user?.email === 'error@inkomi.dev') {
    setStatus(source, 'failed', { errorMsg: 'mock: conversion failed' });
    return finishFolder(parent);
  }
  if (parent?.join) {
    setStatus(source, 'joined');
    return finishFolder(parent);
  }
  if (source.kepubify) {
    setStatus(source, 'kepubifying');
    await delay();
  }
  source.size ??= 1_000_000 + Math.floor(Math.random() * 50_000_000);
  if (!isEpub) source.filename = source.filename.replace(/\.(cbz|pdf)$/i, '.epub');
  setStatus(source, 'done');
  await delay();
  setStatus(source, 'sent');
}

/** A joined folder gets a result once every item settled, then the folder closes. */
async function finishFolder(folder) {
  if (!folder?.join) return;
  const items = childrenOf(folder.id).filter((c) => !c.isResult);
  if (!items.every((i) => i.status === 'joined' || terminal(i.status))) return;
  if (childrenOf(folder.id).some((c) => c.isResult)) return;

  const ok = items.filter((i) => i.status === 'joined');
  if (ok.length === 0) return setStatus(folder, 'failed', { errorMsg: 'mock: every item failed' });

  const result = row(folder.userId, {
    title: folder.title,
    filename: `${folder.title}.epub`,
    parentId: folder.id,
    isResult: true,
    status: 'joining',
    kepubify: folder.kepubify,
    createdAt: folder.createdAt,
  });
  sources.set(result.id, result);
  const { userId, ...resultRow } = result;
  send(userId, { id: result.id, status: 'joining', result: resultRow });

  await delay();
  if (result.kepubify) {
    setStatus(result, 'kepubifying');
    await delay();
  }
  result.size = ok.length * 40_000_000;
  setStatus(result, 'done');
  await delay();
  setStatus(result, 'sent');
  setStatus(folder, 'done');
}

// ─── http ────────────────────────────────────────────────────────────────────────────────

function json(res, status, body) {
  res.writeHead(status, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(body));
}

const fail = (res, status, error) => json(res, status, { error });

function readBody(req) {
  return new Promise((resolve) => {
    const chunks = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', () => resolve(Buffer.concat(chunks)));
  });
}

async function readJson(req) {
  const raw = (await readBody(req)).toString();
  try {
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

/** Like the JWT middleware: 401 in plain text. */
function authUser(req, res) {
  const [scheme, token] = (req.headers.authorization ?? '').split(' ');
  const userId = scheme?.toLowerCase() === 'bearer' ? accessTokens.get(token) : undefined;
  if (!userId) {
    res.writeHead(401, { 'Content-Type': 'text/plain' });
    res.end(
      req.headers.authorization ? 'invalid or expired token' : 'missing authorization header'
    );
    return null;
  }
  return users.get(userId);
}

const routes = [];
const route = (method, pattern, handler) => routes.push({ method, pattern, handler });

route('POST', /^\/auth\/(login|register)$/, async (req, res, [, kind]) => {
  const { email = '', password = '' } = await readJson(req);
  if (!email.includes('@')) return fail(res, 400, 'invalid email');
  if (password.length < 8 && kind === 'register')
    return fail(res, 400, 'password must have between 8 and 128 characters');
  if (!password) return fail(res, 401, 'invalid email or password');
  json(res, 200, issueTokens(userFor(email.toLowerCase()).id));
});

route('POST', /^\/auth\/(google|dropbox)$/, async (req, res, [, provider]) => {
  const user = userFor(`${provider}@inkomi.dev`);
  if (!user.providers.includes(provider)) user.providers.push(provider);
  if (provider === 'dropbox')
    user.dropbox = { linked: true, email: 'usuario@dropbox.com', folder: user.dropbox.folder };
  json(res, 200, issueTokens(user.id));
});

route('POST', /^\/auth\/refresh$/, async (req, res) => {
  const { refreshToken } = await readJson(req);
  const userId = refreshTokens.get(refreshToken);
  if (!userId) return fail(res, 401, 'invalid or expired refresh token');
  refreshTokens.delete(refreshToken);
  json(res, 200, issueTokens(userId));
});

route('POST', /^\/auth\/logout$/, async (req, res) => {
  const { refreshToken } = await readJson(req);
  refreshTokens.delete(refreshToken);
  res.writeHead(204).end();
});

route('GET', /^\/users\/me$/, (req, res) => {
  const user = authUser(req, res);
  if (!user) return;
  json(res, 200, user);
});

route('PUT', /^\/users\/me\/dropbox$/, async (req, res) => {
  const user = authUser(req, res);
  if (!user) return;
  user.dropbox = { ...user.dropbox, linked: true, email: 'usuario@dropbox.com' };
  res.writeHead(204).end();
});

route('DELETE', /^\/users\/me\/dropbox$/, (req, res) => {
  const user = authUser(req, res);
  if (!user) return;
  user.dropbox = { linked: false, folder: user.dropbox.folder };
  res.writeHead(204).end();
});

route('PUT', /^\/users\/me\/dropbox\/folder$/, async (req, res) => {
  const user = authUser(req, res);
  if (!user) return;
  const { folder = '' } = await readJson(req);
  if (!folder.startsWith('/')) return fail(res, 400, 'invalid dropbox folder');
  user.dropbox.folder = folder;
  json(res, 200, { folder });
});

route('PUT', /^\/users\/me\/fcm-token$/, (req, res) => {
  if (!authUser(req, res)) return;
  res.writeHead(204).end();
});

route('GET', /^\/books\/search$/, (req, res, _, url) => {
  if (!authUser(req, res)) return;
  const q = (url.searchParams.get('q') ?? '').trim().toLowerCase();
  if (!q) return fail(res, 400, 'empty query');
  const formats = (url.searchParams.get('fmt') ?? '').split(',').filter(Boolean);
  const words = q.split(/\s+/);
  const found = books.filter(
    (b) =>
      (formats.length === 0 || formats.includes(b.extension)) &&
      words.some((w) => `${b.title} ${b.author}`.toLowerCase().includes(w))
  );
  setTimeout(() => json(res, 200, found), 400);
});

route('POST', /^\/uploads\/request$/, async (req, res) => {
  const user = authUser(req, res);
  if (!user) return;
  const dto = await readJson(req);
  const kepubify = /^Ko/.test(dto.ereader_key ?? '');
  const toRow = (s, extra) =>
    row(user.id, {
      title: s.title,
      filename: s.filename,
      kepubify,
      toCloud: s.to_cloud,
      isRtl: s.reading_direction === 'rtl',
      size: s.size ?? null,
      isFile: s.type !== 'folder',
      ...extra,
    });

  // snake_case back, with the ids and the presigned urls of the files.
  const response = (dto.sources ?? []).map((source) => {
    const root = toRow(source, { join: source.should_join });
    sources.set(root.id, root);
    const items = (source.items ?? []).map((item, position) => {
      const child = toRow(item, { parentId: root.id, position, join: source.should_join });
      sources.set(child.id, child);
      return {
        ...item,
        id: child.id,
        url: item.type === 'file' ? `${origin}/put/${child.id}` : undefined,
      };
    });
    return {
      ...source,
      id: root.id,
      url: source.type === 'file' ? `${origin}/put/${root.id}` : undefined,
      items,
    };
  });
  json(res, 200, response);
});

// The presigned PUT (R2 in production): needs CORS like the bucket.
route('PUT', /^\/put\/([\w-]+)$/, async (req, res, [, id]) => {
  const body = await readBody(req);
  const source = sources.get(id);
  if (!source) return res.writeHead(404).end();
  source.size = body.length;
  source.uploaded = true;
  res.writeHead(200).end();
});

route('PATCH', /^\/uploads\/done\/([\w-]+)$/, (req, res, [, id]) => {
  const user = authUser(req, res);
  if (!user) return;
  const source = sources.get(id);
  if (!source || source.userId !== user.id) return fail(res, 404, 'source not found');
  if (!source.isFile) return fail(res, 400, 'invalid format');
  if (source.status !== 'pending_upload') return res.writeHead(204).end();
  setStatus(source, 'queued');
  res.writeHead(204).end();
  void runSource(source);
});

route('GET', /^\/uploads$/, (req, res, _, url) => {
  const user = authUser(req, res);
  if (!user) return;
  const since = url.searchParams.get('updatedSince');
  const strip = ({ userId, uploaded, ...r }) => r;
  const roots = [...sources.values()]
    .filter((s) => s.userId === user.id && !s.parentId)
    .map((root) => ({ root, children: childrenOf(root.id) }))
    .filter(({ root, children }) => !since || [root, ...children].some((s) => s.updatedAt > since))
    .sort((a, b) => (a.root.createdAt < b.root.createdAt ? 1 : -1))
    .map(({ root, children }) => ({ ...strip(root), children: children.map(strip) }));
  json(res, 200, { uploads: roots, nextCursor: null, syncedAt: now() });
});

route('GET', /^\/uploads\/([\w-]+)\/download$/, (req, res, [, id]) => {
  const user = authUser(req, res);
  if (!user) return;
  const source = sources.get(id);
  if (!source || source.userId !== user.id) return fail(res, 404, 'source not found');
  if (!['done', 'sent'].includes(source.status)) return fail(res, 409, 'source not ready');
  json(res, 200, {
    url: `${origin}/files/${id}`,
    filename: source.filename,
    size: source.size,
    expiresAt: new Date(Date.now() + 15 * 60_000).toISOString(),
  });
});

route('GET', /^\/files\/([\w-]+)$/, (req, res, [, id]) => {
  const source = sources.get(id);
  if (!source) return res.writeHead(404).end();
  res.writeHead(200, {
    'Content-Type': 'application/octet-stream',
    'Content-Disposition': `attachment; filename="${encodeURIComponent(source.filename)}"`,
  });
  res.end(`Inkomi mock file: ${source.filename}\n`);
});

route('GET', /^\/app\/latest$/, (req, res) => fail(res, 404, 'no release available'));

const server = http.createServer(async (req, res) => {
  const url = new URL(req.url, origin);
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, PATCH, DELETE');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');
  if (req.method === 'OPTIONS') return res.writeHead(204).end();

  for (const { method, pattern, handler } of routes) {
    const match = url.pathname.match(pattern);
    if (match && method === req.method) {
      console.log(req.method, url.pathname);
      return handler(req, res, match, url);
    }
  }
  fail(res, 404, 'not found');
});

// ─── websocket ───────────────────────────────────────────────────────────────────────────

const wss = new WebSocketServer({
  noServer: true,
  handleProtocols: (protocols) => (protocols.has('bearer') ? 'bearer' : false),
});

server.on('upgrade', (req, socket, head) => {
  const protocols = (req.headers['sec-websocket-protocol'] ?? '').split(',').map((p) => p.trim());
  const header = (req.headers.authorization ?? '').split(' ')[1];
  const token = protocols[0] === 'bearer' ? protocols[1] : header;
  const userId = accessTokens.get(token);
  if (!userId || !req.url.startsWith('/ws')) {
    socket.write('HTTP/1.1 401 Unauthorized\r\n\r\n');
    return socket.destroy();
  }
  wss.handleUpgrade(req, socket, head, (ws) => {
    sockets.get(userId)?.close();
    sockets.set(userId, ws);
    ws.on('close', () => sockets.get(userId) === ws && sockets.delete(userId));
    console.log('WS open, user', userId);
  });
});

server.listen(port, () => {
  console.log(`Inkomi mock backend on ${origin}`);
  console.log('Web: npm run dev:mock (in another terminal)');
});
