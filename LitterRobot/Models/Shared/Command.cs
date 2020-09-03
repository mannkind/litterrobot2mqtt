using TwoMQTT.Core.Models;

namespace LitterRobot.Models.Shared
{
    /// <summary>
    /// The shared command across the application
    /// </summary>
    public record Command : SharedCommand<Resource>
    {
    }
}
