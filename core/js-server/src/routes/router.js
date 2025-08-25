import { Router } from "express";

import * as relayController from "../controllers/relay.js";
import * as modController from "../controllers/mod.js";
import * as nodeController from "../controllers/node.js";
import { authMiddleware } from "../middlewares/auth.js";

const router = Router();

router.get("/getrelay", relayController.getRelay);
router.post("/postrelay",authMiddleware, relayController.addRelay);
router.delete("/deleterelay",authMiddleware, relayController.deleteRelay);

router.get("/getmod", modController.getMod);
router.post("/postmod",authMiddleware, modController.addMod);
router.delete("/deletemod",authMiddleware, modController.deleteMod);

router.get("/getboot", nodeController.getNode);
router.post("/postboot", authMiddleware,nodeController.addNode);
router.delete("/deleteboot", authMiddleware,nodeController.deleteNode);

export { router };