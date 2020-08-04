using TwoMQTT.Core.Models;

namespace LitterRobot.Models.Options
{
    /// <summary>
    /// The sink options
    /// </summary>
    public class MQTTOpts : MQTTManagerOptions
    {
        public const string Section = "LitterRobot:MQTT";
        public const string TopicPrefixDefault = "home/litterrobot";
        public const string DiscoveryNameDefault = "litterrobot";
    }
}
