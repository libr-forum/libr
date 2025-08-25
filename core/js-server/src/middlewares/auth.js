export const authMiddleware = (req, res, next) => {
    const apiKey = req.headers["x-api-key"];
    console.log("API Key:", apiKey); // Debugging line
    const expectedApiKey = process.env.API_KEY;

    if (!apiKey) {
        return res.status(401).json({ message: "Unauthorized: API key missing" });
    }
    if (apiKey !== expectedApiKey) {
        return res.status(403).json({ message: "Forbidden: Invalid API key" });
    }
    next();
};