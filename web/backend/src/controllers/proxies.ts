import express from 'express';
import lmdb from 'lmdb';
import { isAuthorized, PERMISSIONS } from '../authManager';

export const proxiesRouter = express.Router();

const proxiesDB = lmdb.open<string, string>({ path: `${process.env.DB_PATH}proxies` });
proxiesRouter.post('/api/proxies', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.PROXIES, PERMISSIONS.ACTIONS.POST)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }
  try {
    const bodydata = req.body;
    proxiesDB.put(bodydata.NodeAddr, bodydata.Proxy);
    console.log('put', bodydata.NodeAddr, '=>', bodydata.Proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

proxiesRouter.put('/api/proxies/:nodeAddr', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.PROXIES, PERMISSIONS.ACTIONS.PUT)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }

  let { nodeAddr } = req.params;

  nodeAddr = decodeURIComponent(nodeAddr);

  const proxy = proxiesDB.get(nodeAddr);

  if (proxy === undefined) {
    res.status(404).send('not found');
    return;
  }
  try {
    const bodydata = req.body;
    if (bodydata.Proxy === undefined) {
      res.status(400).send('bad request, proxy is undefined');
      return;
    }

    const { NewNode } = bodydata.NewNode;
    if (NewNode !== nodeAddr) {
      proxiesDB.remove(nodeAddr);
      proxiesDB.put(NewNode, bodydata.Proxy);
    } else {
      proxiesDB.put(nodeAddr, bodydata.Proxy);
    }
    console.log('put', nodeAddr, '=>', bodydata.Proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

proxiesRouter.delete('/api/proxies/:nodeAddr', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.PROXIES, PERMISSIONS.ACTIONS.DELETE)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }

  let { nodeAddr } = req.params;

  nodeAddr = decodeURIComponent(nodeAddr);

  const proxy = proxiesDB.get(nodeAddr);

  if (proxy === undefined) {
    res.status(404).send('not found');
    return;
  }

  try {
    proxiesDB.remove(nodeAddr);
    console.log('remove', nodeAddr, '=>', proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

proxiesRouter.get('/api/proxies', (req, res) => {
  const output = new Map<string, string>();
  proxiesDB.getRange({}).forEach((entry) => {
    output.set(entry.key, entry.value);
  });

  res.status(200).json({ Proxies: Object.fromEntries(output) });
});

proxiesRouter.get('/api/proxies/:nodeAddr', (req, res) => {
  const { nodeAddr } = req.params;

  const proxy = proxiesDB.get(decodeURIComponent(nodeAddr));

  if (proxy === undefined) {
    res.status(404).send('not found');
    return;
  }

  res.status(200).json({
    NodeAddr: nodeAddr,
    Proxy: proxy,
  });
});
