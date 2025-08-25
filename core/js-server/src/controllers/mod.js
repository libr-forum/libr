import { Mods } from "../models/mod.js";

export const getMod = async (req, res) => {
    try {
        const mods = await Mods.find();
        res.json(mods);
    } catch (error) {
        console.error("Error fetching mods:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};

export const addMod = async (req, res) => {
    try {
        const { peer_id, public_key } = req.body;

        if (!peer_id || !public_key) {
            return res.status(400).json({ message: "peer_id and public_key required" });
        }

        // atomic upsert
        const result = await Mods.updateOne(
            { public_key },
            { $set: { peer_id } },
            { upsert: true }
        );

        if (result.upsertedCount > 0) {
            console.log("Mod inserted successfully");
            res.json({ message: "Mod inserted successfully" });
        } else {
            console.log("Mod updated successfully");
            res.json({ message: "Mod updated successfully" });
        }
    } catch (error) {
        console.error("Error adding/updating mod:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};

export const deleteMod = async (req, res) => {
    try {
        const { public_key } = req.body;
        if (!public_key) {
            return res.status(400).json({ message: "public_key required" });
        }

        const result = await Mods.deleteOne({ public_key });

        if (result.deletedCount === 0) {
            return res.status(404).json({ message: "Mod not found" });
        }

        console.log("Mod deleted successfully");
        res.json({ message: "Mod deleted successfully" });
    } catch (error) {
        console.error("Error deleting mod:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};
