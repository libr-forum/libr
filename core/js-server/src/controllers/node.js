import { Nodes } from "../models/node.js";

export const getNode = async (req, res) => {
    try {
        const nodes = await Nodes.find();
        res.json(nodes);
    } catch (error) {
        console.error("Error fetching nodes:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};

export const addNode = async (req, res) => {
    try {
        const { peer_id, node_id } = req.body;

        if (!peer_id || !node_id) {
            return res.status(400).json({ message: "peer_id and node_id required" });
        }

        // atomic upsert
        const result = await Nodes.updateOne(
            { node_id },
            { $set: { peer_id } },
            { upsert: true }
        );

        if (result.upsertedCount > 0) {
            console.log("Node inserted successfully");
            res.json({ message: "Node inserted successfully" });
        } else {
            console.log("Node updated successfully");
            res.json({ message: "Node updated successfully" });
        }
    } catch (error) {
        console.error("Error adding/updating node:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};
