using System;

namespace LitterRobot.Models.Options
{
    /// <summary>
    /// The source options
    /// </summary>
    public record SourceOpts
    {
        public const string Section = "LitterRobot";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string Login { get; init; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string Password { get; init; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <returns></returns>
        public TimeSpan PollingInterval { get; init; } = new TimeSpan(0, 0, 31);
    }
}
