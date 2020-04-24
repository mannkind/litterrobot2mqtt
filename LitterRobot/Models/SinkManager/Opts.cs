using TwoMQTT.Core.Models;

namespace LitterRobot.Models.SinkManager
{
    /// <summary>
    /// The sink options
    /// </summary>
    public class Opts : MQTTManagerOptions
    {
        public const string Section = "LitterRobot";

        /// <summary>
        /// 
        /// </summary>
        public Opts()
        {
            this.TopicPrefix = "home/litterrobot";
            this.DiscoveryName = "litterrobot";
        }
    }
}
