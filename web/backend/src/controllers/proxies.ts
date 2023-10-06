import express from 'express';
import lmdb from 'lmdb';
import { isAuthorized, PERMISSIONS } from '../authManager';

export const proxiesRouter = express.Router();

const proxiesDB = lmdb.open<string, string>({ path: `${process.env.DB_PATH}proxies` });
proxiesRouter.post('', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.PROXIES, PERMISSIONS.ACTIONS.POST)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }
  try {
    const bodydata = req.body;
    proxiesDB.put(bodydata.NodeAddr, bodydata.Proxy);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

proxiesRouter.put('/:nodeAddr', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.PROXIES, PERMISSIONS.ACTIONS.PUT)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }

  let { nodeAddr } = req.params;

  nodeAddr = decodeURIComponent(nodeAddr);

  const proxy = proxiesDB.get(nodeAddr);

  if (proxy === undefined) {
    res.status(404).send(`proxy ${nodeAddr} not found`);
    return;
  }
  try {
    const bodydata = req.body;
    if (bodydata.Proxy === undefined) {
      res.status(400).send(`bad request, proxy ${nodeAddr} is undefined`);
      return;
    }

    const { NewNode } = bodydata;
    if (NewNode !== undefined && NewNode !== nodeAddr) {
      proxiesDB.remove(nodeAddr);
      proxiesDB.put(NewNode, bodydata.Proxy);
    } else {
      proxiesDB.put(nodeAddr, bodydata.Proxy);
    }
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

proxiesRouter.delete('/:nodeAddr', (req, res) => {
  if (!isAuthorized(req.session.userId, PERMISSIONS.SUBJECTS.PROXIES, PERMISSIONS.ACTIONS.DELETE)) {
    res.status(400).send('Unauthorized - only admins and operators allowed');
    return;
  }

  let { nodeAddr } = req.params;

  nodeAddr = decodeURIComponent(nodeAddr);

  const proxy = proxiesDB.get(nodeAddr);

  if (proxy === undefined) {
    res.status(404).send(`proxy ${nodeAddr} not found`);
    return;
  }

  try {
    proxiesDB.remove(nodeAddr);
    res.status(200).send('ok');
  } catch (error: any) {
    res.status(500).send(error.toString());
  }
});

proxiesRouter.get('', (req, res) => {
  const output = new Map<string, string>();
  proxiesDB.getRange({}).forEach((entry) => {
    output.set(entry.key, entry.value);
  });

  res.status(200).json({ Proxies: Object.fromEntries(output) });
});

proxiesRouter.get('/:nodeAddr', (req, res) => {
  const { nodeAddr } = req.params;

  const proxy = proxiesDB.get(decodeURIComponent(nodeAddr));

  if (proxy === undefined) {
    res.status(404).send(`proxy ${nodeAddr} not found`);
    return;
  }

  res.status(200).json({
    NodeAddr: nodeAddr,
    Proxy: proxy,
  });
});
