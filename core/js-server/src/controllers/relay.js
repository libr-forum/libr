import Relay from "../models/relay.js";

export const getRelay = async (req, res) => {
    try {
        const relays = await Relay.find();
        res.json(relays);
    } catch (error) {
        console.error("Error fetching relays:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};

export const addRelay = async (req, res) => {
    try {
        const { address } = req.body;

        if (!address) {
            return res.status(400).json({ message: "address required" });
        }

        const relay = new Relay({ address });
        await relay.save();

        console.log("Relay inserted successfully");
        res.json({ message: "Relay inserted successfully" });
    } catch (error) {
        console.error("Error adding relay:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};

export const deleteRelay = async (req, res) => {
    try {
        const { address } = req.body;
        if (!address) {
            return res.status(400).json({ message: "address required" });
        }

        const result = await Relay.deleteOne({ address });

        if (result.deletedCount === 0) {
            return res.status(404).json({ message: "Relay not found" });
        }

        console.log("Relay deleted successfully");
        res.json({ message: "Relay deleted successfully" });
    } catch (error) {
        console.error("Error deleting relay:", error);
        res.status(500).json({ message: "Internal server error" });
    }
};
