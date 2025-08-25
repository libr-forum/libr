import { Nodes } from "../models/node.js";

export const getNode = async (req, res) => {
    try {
        const nodes = await Nodes.find();
        res.json({ boot_list: nodes });  // wrap inside boot_list
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

export const deleteNode = async (req, res) => {
    try {
        const { node_id } = req.body;
        if (!node_id) {
            return res.status(400).json({ message: "node_id required" });
        }

        const result = await Nodes.deleteOne({ node_id });

        if (result.deletedCount === 0) {
            return res.status(404).json({ message: "Node not found" });
        }

        console.log("Node deleted successfully");
        res.json({ message: "Node deleted successfully" });
    } catch (error) {
        console.error("Error deleting node:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};
