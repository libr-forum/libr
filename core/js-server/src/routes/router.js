import { Router } from "express";

import * as relayController from "../controllers/relay.js";
import * as modController from "../controllers/mod.js";
import * as nodeController from "../controllers/node.js";

const router = Router();

router.get("/getrelay", relayController.getRelay);
router.post("/postrelay", relayController.addRelay);
router.delete("/deleterelay", relayController.deleteRelay);

router.get("/getmod", modController.getMod);
router.post("/postmod", modController.addMod);
router.delete("/deletemod", modController.deleteMod);

router.get("/getboot", nodeController.getNode);
router.post("/postboot", nodeController.addNode);
router.delete("/deleteboot", nodeController.deleteNode);

export { router };